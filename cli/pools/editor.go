package pools

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/cli/form"
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/cli/widget"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/pool"
	"github.com/pritunl/pritunl-cloud/zone"
	"github.com/sirupsen/logrus"
)

// commitFields are the fields written by the admin pool handler.
var commitFields = set.NewSet(
	"name",
	"comment",
	"delete_protection",
	"type",
)

// editor is the pool settings form mirroring the inputs of the detailed
// pool view in the web interface, the zone and volume group of an
// existing pool cannot be changed.
type editor struct {
	pl     *pool.Pool
	names  *names
	create bool
	form   *form.Form

	name             *form.Text
	comment          *form.Area
	datacenter       *form.Select
	zone             *form.Select
	vgName           *form.Text
	deleteProtection *form.Toggle
}

func newEditor(pl *pool.Pool, nms *names) *editor {
	e := &editor{
		pl:    pl,
		names: nms,
	}

	e.name = form.NewText("Name", "Enter name", pl.Name)
	e.comment = form.NewArea("Comment", "Pool comment", pl.Comment, 4)

	// The datacenter only filters the zones of a new pool, the first
	// datacenter is the default like the web interface
	dcOpts := resource.SelectOptions(nms.datacenters)
	if len(dcOpts) == 0 {
		dcOpts = resource.WithSelect("No Datacenters", nil)
	}
	e.datacenter = form.NewSelect("Datacenter", dcOpts,
		resource.IdHex(pl.Datacenter))
	e.datacenter.Hide = func() bool {
		return !pl.Zone.IsZero()
	}
	e.datacenter.OnChange = func() {
		e.zone.SetValue("")
	}

	e.zone = form.NewSelect("Zone", nil, resource.IdHex(pl.Zone))
	e.zone.Dynamic = e.zoneOptions
	e.zone.Lock = func() bool {
		return !pl.Zone.IsZero()
	}

	e.vgName = form.NewText("Volume Group Name", "Enter name", pl.VgName)
	e.vgName.Lock = func() bool {
		return !e.create
	}

	e.deleteProtection = form.NewToggle("Delete protection",
		pl.DeleteProtection)

	e.form = form.New(
		e.name,
		e.comment,
		e.datacenter,
		e.zone,
		e.vgName,
		e.deleteProtection,
	)

	return e
}

// zoneOptions lists the zones of the selected datacenter for a new
// pool, an existing pool lists every zone like the web interface.
func (e *editor) zoneOptions() []widget.SelectOption {
	dcId := resource.ParseId(e.datacenter.Value())
	opts := []widget.SelectOption{}
	for _, zne := range e.names.zones {
		if e.pl.Zone.IsZero() && zne.Datacenter != dcId {
			continue
		}
		opts = append(opts, widget.SelectOption{
			Label: zne.Name,
			Value: zne.Id.Hex(),
		})
	}
	if len(opts) == 0 {
		return resource.WithSelect("No Zones", nil)
	}
	return resource.WithSelect("Select Zone", opts)
}

func (e *editor) Form() *form.Form {
	return e.form
}

// createZone returns the zone of a new pool, the last zone of the
// datacenter when none is selected like the web interface.
func (e *editor) createZone() bson.ObjectID {
	zoneId := resource.ParseId(e.zone.Value())
	if !zoneId.IsZero() {
		return zoneId
	}

	dcId := resource.ParseId(e.datacenter.Value())
	for _, zne := range e.names.zones {
		if zne.Datacenter == dcId {
			zoneId = zne.Id
		}
	}
	return zoneId
}

func saveError(message string) error {
	return &errortypes.ParseError{
		errors.New("pools: " + message),
	}
}

// Save applies the form to the current pool the same way as the admin
// pool handler, or inserts a new pool in the selected zone.
func (e *editor) Save(db *database.Database) (err error) {
	var pl *pool.Pool
	if e.create {
		zoneId := e.createZone()
		if zoneId.IsZero() {
			err = saveError("Missing required zone")
			return
		}

		zne, er := zone.Get(db, zoneId)
		if er != nil {
			err = er
			return
		}

		pl = &pool.Pool{
			Datacenter: zne.Datacenter,
			Zone:       zoneId,
			Type:       pool.Lvm,
			VgName:     e.vgName.Value(),
		}
	} else {
		pl, err = pool.Get(db, e.pl.Id)
		if err != nil {
			return
		}
	}

	pl.Name = e.name.Value()
	pl.Comment = e.comment.Value()
	pl.DeleteProtection = e.deleteProtection.Value()

	errData, err := pl.Validate(db)
	if err != nil {
		return
	}
	if errData != nil {
		err = saveError(errData.Message)
		return
	}

	if e.create {
		err = pl.Insert(db)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"pool_id": pl.Id.Hex(),
		}).Info("tui: Pool created")
	} else {
		err = pl.CommitFields(db, commitFields)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"pool_id": pl.Id.Hex(),
		}).Info("tui: Pool settings saved")
	}

	err = event.PublishDispatch(db, "pool.change")
	if err != nil {
		return
	}

	return
}

package shapes

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
	"github.com/pritunl/pritunl-cloud/shape"
	"github.com/sirupsen/logrus"
)

// commitFields are the fields written by the admin shape handler.
var commitFields = set.NewSet(
	"name",
	"type",
	"comment",
	"delete_protection",
	"roles",
	"flexible",
	"disk_type",
	"disks_path",
	"disk_pools",
	"memory",
	"processors",
)

var typeOptions = []widget.SelectOption{
	{Label: "Instance", Value: shape.Instance},
	{Label: "Disk", Value: shape.Disk},
}

var diskTypeOptions = []widget.SelectOption{
	{Label: "QCOW", Value: shape.Qcow2},
	{Label: "LVM", Value: shape.Lvm},
}

// editor is the shape settings form mirroring the inputs of the
// detailed shape view in the web interface. Instance shapes have the
// processor and memory inputs, disk shapes have a disk type that picks
// between a disks path and the disk pools, the disk type is not stored.
type editor struct {
	shpe   *shape.Shape
	pools  []widget.SelectOption
	create bool
	form   *form.Form

	name             *form.Text
	comment          *form.Area
	typ              *form.Select
	diskType         *form.Select
	disksPath        *form.Text
	diskPools        *form.Tokens
	flexible         *form.Toggle
	deleteProtection *form.Toggle
	memory           *form.Number
	processors       *form.Number
	roles            *form.Tokens
}

func newEditor(shpe *shape.Shape, pools []widget.SelectOption) *editor {
	e := &editor{
		shpe:  shpe,
		pools: pools,
	}

	diskType := shpe.DiskType
	if diskType == "" {
		diskType = shape.Qcow2
	}

	diskPools := []string{}
	for _, poolId := range shpe.DiskPools {
		diskPools = append(diskPools, poolId.Hex())
	}

	e.name = form.NewText("Name", "Enter name", shpe.Name)
	e.comment = form.NewArea("Comment", "Shape comment", shpe.Comment, 4)
	e.typ = form.NewSelect("Type", typeOptions, shpe.GetType())

	e.diskType = form.NewSelect("Disk Type", diskTypeOptions, diskType)
	e.diskType.Hide = func() bool {
		return !e.isDisk()
	}

	e.disksPath = form.NewText("Disks Path", "/var/lib/pritunl-cloud/disks",
		shpe.DisksPath)
	e.disksPath.Hide = func() bool {
		return !e.isDisk() || e.diskType.Value() != shape.Qcow2
	}

	e.diskPools = form.NewTokensSelect("Disk Pools", pools, diskPools)
	e.diskPools.Placeholder = "No pools"
	e.diskPools.Hide = func() bool {
		return !e.isDisk() || e.diskType.Value() != shape.Lvm
	}

	e.flexible = form.NewToggle("Flexible", shpe.Flexible)
	e.flexible.Hide = e.isDisk
	e.deleteProtection = form.NewToggle("Delete protection",
		shpe.DeleteProtection)
	e.memory = form.NewNumber("Memory Size (MB)", "Memory in megabytes",
		shpe.Memory, 256, 0, 1024)
	e.memory.Hide = e.isDisk
	e.processors = form.NewNumber("Processors", "Number of processors",
		shpe.Processors, 1, 0, 1)
	e.processors.Hide = e.isDisk
	e.roles = form.NewTokens("Roles", "Add role", shpe.Roles)

	e.form = form.New(
		e.name,
		e.comment,
		e.typ,
		e.diskType,
		e.disksPath,
		e.diskPools,
		e.flexible,
		e.deleteProtection,
		e.memory,
		e.processors,
		e.roles,
	)

	return e
}

func (e *editor) isDisk() bool {
	return e.typ.Value() == shape.Disk
}

func (e *editor) Form() *form.Form {
	return e.form
}

// apply sets the form values on the shape.
func (e *editor) apply(shpe *shape.Shape) {
	shpe.Name = e.name.Value()
	shpe.Comment = e.comment.Value()
	shpe.Type = e.typ.Value()
	shpe.DeleteProtection = e.deleteProtection.Value()
	shpe.Roles = e.roles.Values()
	shpe.Flexible = e.flexible.Value()
	shpe.Memory = e.memory.Int()
	shpe.Processors = e.processors.Int()

	shpe.DiskType = ""
	shpe.DisksPath = ""
	shpe.DiskPools = []bson.ObjectID{}
	if e.isDisk() {
		shpe.DiskType = e.diskType.Value()
		switch shpe.DiskType {
		case shape.Lvm:
			for _, poolId := range e.diskPools.Values() {
				shpe.DiskPools = append(shpe.DiskPools,
					resource.ParseId(poolId))
			}
		default:
			shpe.DisksPath = e.disksPath.Value()
		}
	}
}

// Save applies the form to the current shape the same way as the admin
// shape handler, or inserts a new shape.
func (e *editor) Save(db *database.Database) (err error) {
	var shpe *shape.Shape
	if e.create {
		shpe = &shape.Shape{}
	} else {
		shpe, err = shape.Get(db, e.shpe.Id)
		if err != nil {
			return
		}
	}

	e.apply(shpe)

	errData, err := shpe.Validate(db)
	if err != nil {
		return
	}
	if errData != nil {
		err = &errortypes.ParseError{
			errors.New("shapes: " + errData.Message),
		}
		return
	}

	if e.create {
		err = shpe.Insert(db)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"shape_id": shpe.Id.Hex(),
		}).Info("tui: Shape created")
	} else {
		err = shpe.CommitFields(db, commitFields)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"shape_id": shpe.Id.Hex(),
		}).Info("tui: Shape settings saved")
	}

	err = event.PublishDispatch(db, "shape.change")
	if err != nil {
		return
	}

	return
}

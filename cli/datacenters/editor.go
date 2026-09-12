package datacenters

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/cli/form"
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/cli/widget"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/datacenter"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/storage"
	"github.com/sirupsen/logrus"
)

// commitFields are the fields written by the admin datacenter handler.
var commitFields = set.NewSet(
	"name",
	"comment",
	"network_mode",
	"match_organizations",
	"organizations",
	"jumbo_mtu",
	"public_storages",
	"private_storage",
	"private_storage_class",
	"backup_storage",
	"backup_storage_class",
)

var storageClassOptions = []widget.SelectOption{
	{Label: "Default", Value: ""},
	{Label: "AWS Standard", Value: "aws_standard"},
	{Label: "AWS Standard-IA", Value: "aws_infrequent_access"},
	{Label: "AWS Glacier", Value: "aws_glacier"},
}

var networkModeOptions = []widget.SelectOption{
	{Label: "Default", Value: datacenter.Default},
	{Label: "VXLAN", Value: datacenter.VxlanVlan},
}

// editor is the datacenter settings form mirroring the inputs of the
// detailed datacenter view in the web interface.
type editor struct {
	dc     *datacenter.Datacenter
	create bool
	form   *form.Form

	name                *form.Text
	comment             *form.Area
	privateStorage      *form.Select
	privateStorageClass *form.Select
	backupStorage       *form.Select
	backupStorageClass  *form.Select
	publicStorages      *form.Tokens
	matchOrganizations  *form.Toggle
	organizations       *form.Tokens
	networkMode         *form.Select
	jumboMtu            *form.Number
}

func hexList(ids []bson.ObjectID) []string {
	hexes := []string{}
	for _, id := range ids {
		hexes = append(hexes, id.Hex())
	}
	return hexes
}

func idList(hexes []string) []bson.ObjectID {
	ids := []bson.ObjectID{}
	for _, hex := range hexes {
		id := resource.ParseId(hex)
		if !id.IsZero() {
			ids = append(ids, id)
		}
	}
	return ids
}

func newEditor(dc *datacenter.Datacenter, nms *names) *editor {
	e := &editor{
		dc: dc,
	}

	e.name = form.NewText("Name", "Enter name", dc.Name)
	e.comment = form.NewArea("Comment", "Datacenter comment", dc.Comment, 4)

	// Public and web storages hold images, private storages hold
	// snapshots and backups
	privateOpts := []widget.SelectOption{{Label: "None", Value: ""}}
	publicOpts := []widget.SelectOption{}
	for _, store := range nms.storages {
		opt := widget.SelectOption{
			Label: store.Name,
			Value: store.Id.Hex(),
		}
		switch store.Type {
		case storage.Public, storage.Web:
			publicOpts = append(publicOpts, opt)
		case storage.Private:
			privateOpts = append(privateOpts, opt)
		}
	}

	e.privateStorage = form.NewSelect("Private Storage", privateOpts,
		resource.IdHex(dc.PrivateStorage))
	e.privateStorageClass = form.NewSelect("Private Storage Class",
		storageClassOptions, dc.PrivateStorageClass)
	e.backupStorage = form.NewSelect("Backup Storage", privateOpts,
		resource.IdHex(dc.BackupStorage))
	e.backupStorageClass = form.NewSelect("Backup Storage Class",
		storageClassOptions, dc.BackupStorageClass)

	e.publicStorages = form.NewTokensSelect("Public Storages", publicOpts,
		hexList(dc.PublicStorages))

	e.matchOrganizations = form.NewToggle("Match organizations",
		dc.MatchOrganizations)
	e.organizations = form.NewTokensSelect("Organizations",
		resource.SelectOptions(nms.orgs), hexList(dc.Organizations))
	e.organizations.Hide = func() bool {
		return !e.matchOrganizations.Value()
	}

	e.networkMode = form.NewSelect("Network Mode", networkModeOptions,
		dc.NetworkMode)
	e.jumboMtu = form.NewNumber("Jumbo Frames MTU", "9000", dc.JumboMtu,
		0, 65535, 1)

	e.form = form.New(
		e.name,
		e.comment,
		e.privateStorage,
		e.privateStorageClass,
		e.backupStorage,
		e.backupStorageClass,
		e.publicStorages,
		e.matchOrganizations,
		e.organizations,
		e.networkMode,
		e.jumboMtu,
	)

	return e
}

func (e *editor) Form() *form.Form {
	return e.form
}

// apply sets the form values on the datacenter.
func (e *editor) apply(dc *datacenter.Datacenter) {
	dc.Name = e.name.Value()
	dc.Comment = e.comment.Value()
	dc.NetworkMode = e.networkMode.Value()
	dc.MatchOrganizations = e.matchOrganizations.Value()
	dc.Organizations = idList(e.organizations.Values())
	dc.JumboMtu = e.jumboMtu.Int()
	dc.PublicStorages = idList(e.publicStorages.Values())
	dc.PrivateStorage = resource.ParseId(e.privateStorage.Value())
	dc.PrivateStorageClass = e.privateStorageClass.Value()
	dc.BackupStorage = resource.ParseId(e.backupStorage.Value())
	dc.BackupStorageClass = e.backupStorageClass.Value()
}

// Save applies the form to the current datacenter the same way as the
// admin datacenter handler, or inserts a new datacenter.
func (e *editor) Save(db *database.Database) (err error) {
	var dc *datacenter.Datacenter
	if e.create {
		dc = &datacenter.Datacenter{}
	} else {
		dc, err = datacenter.Get(db, e.dc.Id)
		if err != nil {
			return
		}
	}

	e.apply(dc)

	errData, err := dc.Validate(db)
	if err != nil {
		return
	}
	if errData != nil {
		err = &errortypes.ParseError{
			errors.New("datacenters: " + errData.Message),
		}
		return
	}

	if e.create {
		err = dc.Insert(db)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"datacenter_id": dc.Id.Hex(),
		}).Info("tui: Datacenter created")
	} else {
		err = dc.CommitFields(db, commitFields)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"datacenter_id": dc.Id.Hex(),
		}).Info("tui: Datacenter settings saved")
	}

	err = event.PublishDispatch(db, "datacenter.change")
	if err != nil {
		return
	}

	return
}

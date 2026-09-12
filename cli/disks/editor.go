package disks

import (
	"strconv"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/aggregate"
	"github.com/pritunl/pritunl-cloud/cli/form"
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/cli/widget"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/sirupsen/logrus"
)

var diskTypeOptions = []widget.SelectOption{
	{Label: "QCOW", Value: disk.Qcow2},
	{Label: "LVM", Value: disk.Lvm},
}

// editor is the disk settings form mirroring the inputs of the detailed
// disk view in the web interface.
type editor struct {
	dsk  *aggregate.DiskAggregate
	form *form.Form

	name             *form.Text
	comment          *form.Area
	typ              *form.Select
	instance         *form.Select
	index            *form.Number
	deleteProtection *form.Toggle
	resize           *form.Toggle
	newSize          *form.Number
	backup           *form.Toggle
}

// indexNumber returns the numeric disk index or zero for the hold
// index of detached disks like the web interface.
func indexNumber(index string) int {
	value, err := strconv.Atoi(index)
	if err != nil {
		return 0
	}
	return value
}

func newEditor(dsk *aggregate.DiskAggregate,
	instances []widget.SelectOption) *editor {

	e := &editor{
		dsk: dsk,
	}

	e.name = form.NewText("Name", "Enter name", dsk.Name)
	e.comment = form.NewArea("Comment", "Disk comment", dsk.Comment, 4)

	// The type of an existing disk cannot be changed
	e.typ = form.NewSelect("Type", diskTypeOptions, dsk.Type)
	e.typ.Lock = func() bool {
		return true
	}

	instOpts := resource.WithSelect("No Instances", nil)
	if len(instances) > 0 {
		instOpts = resource.WithSelect("Detached Disk", instances)
	}
	e.instance = form.NewSelect("Instance", instOpts,
		resource.IdHex(dsk.Instance))
	e.instance.Lock = func() bool {
		return len(instances) == 0
	}

	e.index = form.NewNumber("Index", "Disk index", indexNumber(dsk.Index),
		0, 8, 1)
	e.index.Hide = func() bool {
		return e.instance.Value() == ""
	}

	e.deleteProtection = form.NewToggle("Delete protection",
		dsk.DeleteProtection)

	// Resizing copies the current size to the new size like the web
	// interface, the size can only grow
	e.resize = form.NewToggle("Resize disk", false)
	e.resize.Lock = func() bool {
		return dsk.State != disk.Available
	}
	e.newSize = form.NewNumber("New Size (GB)", "New disk size in gigabytes",
		0, dsk.Size, 0, 1)
	e.newSize.Hide = func() bool {
		return !e.resize.Value()
	}
	e.resize.OnChange = func() {
		if e.resize.Value() {
			e.newSize.Set(dsk.Size)
		} else {
			e.newSize.Set(0)
		}
	}

	e.backup = form.NewToggle("Automatic backup", dsk.Backup)

	e.form = form.New(
		e.name,
		e.comment,
		e.typ,
		e.instance,
		e.index,
		e.deleteProtection,
		e.resize,
		e.newSize,
		e.backup,
	)

	return e
}

func (e *editor) Form() *form.Form {
	return e.form
}

// Save applies the form to the current disk the same way as the admin
// disk handler, a resize queues the expand action.
func (e *editor) Save(db *database.Database) (err error) {
	dsk, err := disk.Get(db, e.dsk.Id)
	if err != nil {
		return
	}

	fields := set.NewSet(
		"name",
		"comment",
		"type",
		"instance",
		"delete_protection",
		"index",
		"backup",
		"new_size",
	)

	dsk.PreCommit()
	dsk.Name = e.name.Value()
	dsk.Comment = e.comment.Value()
	dsk.Instance = resource.ParseId(e.instance.Value())
	dsk.DeleteProtection = e.deleteProtection.Value()
	if !dsk.Instance.IsZero() {
		dsk.Index = strconv.Itoa(e.index.Int())
	}
	dsk.Backup = e.backup.Value()

	newSize := e.newSize.Int()
	if e.resize.Value() && newSize > dsk.Size {
		if dsk.Action != "" {
			err = &errortypes.ParseError{
				errors.New("disks: Disk action already active"),
			}
			return
		}

		if dsk.IsActive() {
			dsk.Action = disk.Expand
			dsk.NewSize = newSize
			fields.Add("action")
		}
	}

	errData, err := dsk.Validate(db)
	if err != nil {
		return
	}
	if errData != nil {
		err = &errortypes.ParseError{
			errors.New("disks: " + errData.Message),
		}
		return
	}

	err = dsk.CommitFields(db, fields)
	if err != nil {
		return
	}

	logrus.WithFields(logrus.Fields{
		"disk_id": dsk.Id.Hex(),
	}).Info("tui: Disk settings saved")

	err = event.PublishDispatch(db, "disk.change")
	if err != nil {
		return
	}

	return
}

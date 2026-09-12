package disks

import (
	"fmt"
	"github.com/pritunl/pritunl-cloud/instance"
	"image/color"
	"strings"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/aggregate"
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/cli/widget"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/image"
	"github.com/sirupsen/logrus"
)

// Item is a disk card with the referenced names resolved and the
// instances of its node for the instance select.
type Item struct {
	dsk       *aggregate.DiskAggregate
	instances []widget.SelectOption
	org       string
	node      string
	pool      string
	image     string
}

func (i *Item) Disk() *disk.Disk {
	return &i.dsk.Disk
}

func (i *Item) Id() string {
	return i.dsk.Id.Hex()
}

func (i *Item) Name() string {
	return i.dsk.Name
}

func (i *Item) Tag() string {
	return i.dsk.Id.Hex()
}

// status returns the status text and color of the web disk rows from
// the disk state and action.
func status(dsk *disk.Disk) (text string, clr color.Color) {
	text = "Unknown"

	switch dsk.State {
	case disk.Provision:
		text = "Provisioning"
		clr = widget.ColorAccent
	case disk.Available:
		if !dsk.Instance.IsZero() {
			text = "Connected"
		} else {
			text = "Available"
		}
		clr = widget.ColorGreen
	case disk.Attached:
		text = "Connected"
		clr = widget.ColorGreen
	}

	switch dsk.Action {
	case disk.Destroy:
		text = "Destroying"
		clr = widget.ColorRed
	case disk.Snapshot:
		text = "Snapshotting"
		clr = widget.ColorAccent
	case disk.Backup:
		text = "Backing Up"
		clr = widget.ColorAccent
	case disk.Restore:
		text = "Restoring"
		clr = widget.ColorAccent
	case disk.Expand:
		text = "Expanding"
		clr = widget.ColorAccent
	}

	return
}

// resourceField returns the pool of an LVM disk or the node of a QCOW
// disk like the web disk rows.
func (i *Item) resourceField() (label, value string) {
	if i.dsk.Type == disk.Lvm {
		return "Pool", widget.Default(i.pool, "Pool Unavailable")
	}
	return "Node", widget.Default(i.node, resource.IdHex(i.dsk.Node))
}

func (i *Item) organization() string {
	if i.dsk.Organization.IsZero() {
		return "Unknown Organization"
	}
	return widget.Default(i.org, i.dsk.Organization.Hex())
}

func (i *Item) instanceName() string {
	if i.dsk.InstanceInfo != nil {
		return widget.Default(i.dsk.InstanceInfo.Name,
			i.dsk.InstanceInfo.Id.Hex())
	}
	return resource.IdHex(i.dsk.Instance)
}

func (i *Item) Fields() []resource.Field {
	statusText, statusColor := status(&i.dsk.Disk)
	resourceLabel, resourceValue := i.resourceField()

	return []resource.Field{
		{
			Label: "Status",
			Value: statusText,
			Color: statusColor,
		},
		{
			Label: "Organization",
			Value: i.organization(),
		},
		{
			Label: resourceLabel,
			Value: resourceValue,
		},
		{
			Label: "Instance",
			Value: i.instanceName(),
		},
		{
			Label: "Size",
			Value: fmt.Sprintf("%dGB", i.dsk.Size),
		},
	}
}

// instanceStatusColor matches the status colors of the attached
// instance card in the web interface.
func instanceStatusColor(status string) color.Color {
	switch status {
	case "Running":
		return widget.ColorGreen
	case "Starting", "Restarting", "Updating", "Provisioning":
		return widget.ColorAccent
	case "Failed", "Stopping", "Stopped", "Destroying":
		return widget.ColorRed
	}
	if strings.HasPrefix(status, "Restart Required") {
		return widget.ColorYellow
	}
	return nil
}

// Info mirrors the fields of the detailed disk view followed by the
// attached instance card.
func (i *Item) Info() []widget.InfoField {
	dsk := i.dsk
	resourceLabel, resourceValue := i.resourceField()

	fields := []widget.InfoField{
		{"ID", dsk.Id.Hex()},
		{"Organization", i.organization()},
	}

	if !dsk.Image.IsZero() {
		fields = append(fields, widget.InfoField{"Image",
			widget.Default(i.image, dsk.Image.Hex())})
	}
	if dsk.BackingImage != "" {
		fields = append(fields, widget.InfoField{"Backing Image",
			dsk.BackingImage})
	}
	if dsk.Uuid != "" {
		fields = append(fields, widget.InfoField{"UUID", dsk.Uuid})
	}
	if dsk.FileSystem != "" {
		fields = append(fields, widget.InfoField{"File System",
			dsk.FileSystem})
	}

	fields = append(fields,
		widget.InfoField{resourceLabel, resourceValue},
		widget.InfoField{"Size", fmt.Sprintf("%dGB", dsk.Size)},
	)

	if !dsk.Deployment.IsZero() {
		fields = append(fields, widget.InfoField{"Deployment",
			dsk.Deployment.Hex()})
	}

	inst := dsk.InstanceInfo
	if inst != nil {
		fields = append(fields,
			widget.InfoField{"Instance", inst.Name},
			widget.InfoField{"Instance ID", inst.Id.Hex()},
			widget.InfoField{"Instance Index", dsk.Index},
			widget.InfoField{"Instance Status", inst.Status},
			widget.InfoField{"Instance Uptime", inst.Uptime},
			widget.InfoField{"Private IPv4", strings.Join(inst.PrivateIps, ", ")},
			widget.InfoField{"Public IPv4", strings.Join(inst.PublicIps, ", ")},
			widget.InfoField{"Private IPv6", strings.Join(inst.PrivateIps6, ", ")},
			widget.InfoField{"Public IPv6", strings.Join(inst.PublicIps6, ", ")},
		)
	}

	return fields
}

// setAction queues a snapshot or backup the same way as the bulk
// action of the admin disks handler.
func setAction(diskId bson.ObjectID, action string) func(
	db *database.Database) error {

	return func(db *database.Database) (err error) {
		doc := bson.M{
			"action": action,
		}

		err = disk.UpdateMulti(db, []bson.ObjectID{diskId}, &doc)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"disk_id": diskId.Hex(),
			"action":  action,
		}).Info("tui: Disk action")

		err = event.PublishDispatch(db, "disk.change")
		if err != nil {
			return
		}

		return
	}
}

// restore queues the restore of a backup image the same way as the
// restore of the admin disk handler.
func restore(diskId, imageId bson.ObjectID) func(
	db *database.Database) error {

	return func(db *database.Database) (err error) {
		dsk, err := disk.Get(db, diskId)
		if err != nil {
			return
		}

		if dsk.Action != "" {
			err = &errortypes.ParseError{
				errors.New("disks: Disk action already active"),
			}
			return
		}

		if !dsk.IsActive() {
			err = &errortypes.ParseError{
				errors.New("disks: Disk not available"),
			}
			return
		}

		img, err := image.Get(db, imageId)
		if err != nil {
			return
		}

		if img.Disk != dsk.Id {
			err = &errortypes.ParseError{
				errors.New("disks: Invalid restore image"),
			}
			return
		}

		dsk.PreCommit()
		dsk.Action = disk.Restore
		dsk.RestoreImage = img.Id

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

		err = dsk.CommitFields(db, set.NewSet("action", "restore_image"))
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"disk_id":  dsk.Id.Hex(),
			"image_id": img.Id.Hex(),
		}).Info("tui: Disk restore started")

		err = event.PublishDispatch(db, "disk.change")
		if err != nil {
			return
		}

		return
	}
}

// deleteDisk queues the disk destroy the same way as the admin disk
// handler, disks and instances with delete protection are refused.
func deleteDisk(diskId bson.ObjectID) func(db *database.Database) error {
	return func(db *database.Database) (err error) {
		dsk, err := disk.Get(db, diskId)
		if err != nil {
			return
		}

		if dsk.DeleteProtection {
			err = resource.DeleteError("disk",
				"Cannot delete disk with delete protection")
			return
		}

		if !dsk.Instance.IsZero() {
			inst, e := instance.Get(db, dsk.Instance)
			if e != nil {
				err = e
				return
			}

			if inst.DeleteProtection {
				err = resource.DeleteError("disk", "Cannot delete disk "+
					"attached to instance with delete protection")
				return
			}
		}

		return resource.Delete("disk", "disk.change", diskId, disk.Delete)(db)
	}
}

// Actions mirrors the snapshot, backup, restore and delete buttons of
// the web interface, the disk actions are shown while the disk is
// available with no action queued. Restore uses the most recent backup
// like the default choice of the web restore menu.
func (i *Item) Actions() []resource.Action {
	dsk := i.dsk
	deleteAction := resource.DeleteAction("disk", dsk.Name,
		deleteDisk(dsk.Id))

	if !dsk.IsActive() || dsk.Action != "" {
		return []resource.Action{deleteAction}
	}

	actions := []resource.Action{
		{
			Key:     "s",
			Label:   "Snapshot",
			Status:  "Snapshotting",
			Confirm: fmt.Sprintf("Snapshot the disk %s?", dsk.Name),
			Run:     setAction(dsk.Id, disk.Snapshot),
		},
		{
			Key:     "b",
			Label:   "Backup",
			Status:  "Backing up",
			Confirm: fmt.Sprintf("Backup the disk %s?", dsk.Name),
			Run:     setAction(dsk.Id, disk.Backup),
		},
	}

	if len(dsk.Backups) > 0 {
		backup := dsk.Backups[0]
		actions = append(actions, resource.Action{
			Key:    "r",
			Label:  "Restore",
			Status: "Restoring",
			Danger: true,
			Confirm: fmt.Sprintf("Restore the disk %s from backup %s? "+
				"The instance will be stopped.", dsk.Name, backup.Name),
			Run: restore(dsk.Id, backup.Image),
		})
	}

	return append(actions, deleteAction)
}

func (i *Item) Editor() resource.Editor {
	return newEditor(i.dsk, i.instances)
}

package disk

import (
	"context"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/utils"
	"strconv"
	"strings"
	"time"
)

type Disk struct {
	Id               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name             string             `bson:"name" json:"name"`
	State            string             `bson:"state" json:"state"`
	Node             primitive.ObjectID `bson:"node" json:"node"`
	Organization     primitive.ObjectID `bson:"organization,omitempty" json:"organization"`
	Instance         primitive.ObjectID `bson:"instance,omitempty" json:"instance"`
	SourceInstance   primitive.ObjectID `bson:"source_instance,omitempty" json:"source_instance"`
	DeleteProtection bool               `bson:"delete_protection" json:"delete_protection"`
	Image            primitive.ObjectID `bson:"image,omitempty" json:"image"`
	RestoreImage     primitive.ObjectID `bson:"restore_image,omitempty" json:"restore_image"`
	Backing          bool               `bson:"backing" json:"backing"`
	BackingImage     string             `bson:"backing_image" json:"backing_image"`
	Index            string             `bson:"index" json:"index"`
	Size             int                `bson:"size" json:"size"`
	Backup           bool               `bson:"backup" json:"backup"`
	LastBackup       time.Time          `bson:"last_backup" json:"last_backup"`
}

func (d *Disk) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	if !d.Instance.IsZero() && d.Index != "" {
		index, e := strconv.Atoi(d.Index)
		if e != nil {
			errData = &errortypes.ErrorData{
				Error:   "index_invalid",
				Message: "Disk index invalid",
			}
			return
		}

		if index < 0 || index > 10 {
			errData = &errortypes.ErrorData{
				Error:   "index_out_of_range",
				Message: "Disk index out of range",
			}
			return
		}

		d.Index = strconv.Itoa(index)
	}

	if d.Backup && d.BackingImage != "" {
		errData = &errortypes.ErrorData{
			Error:   "backing_image_backup",
			Message: "Cannot enable backups with backing image",
		}
		return
	}

	if d.Instance.IsZero() && !strings.HasPrefix(d.Index, "hold") {
		d.Index = fmt.Sprintf("hold_%s", primitive.NewObjectID().Hex())
	}

	if d.State == "" {
		d.State = Provision
	}

	if d.Size < 10 {
		d.Size = 10
	}

	return
}

func (d *Disk) Commit(db *database.Database) (err error) {
	coll := db.Disks()

	err = coll.Commit(d.Id, d)
	if err != nil {
		return
	}

	return
}

func (d *Disk) CommitFields(db *database.Database, fields set.Set) (
	err error) {

	coll := db.Disks()

	err = coll.CommitFields(d.Id, d, fields)
	if err != nil {
		return
	}

	return
}

func (d *Disk) Insert(db *database.Database) (err error) {
	coll := db.Disks()

	_, err = coll.InsertOne(context.Background(), d)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func (d *Disk) Destroy(db *database.Database) (err error) {
	dskPath := paths.GetDiskPath(d.Id)

	if d.DeleteProtection {
		logrus.WithFields(logrus.Fields{
			"disk_id": d.Id.Hex(),
		}).Info("disk: Delete protection ignore disk destroy")

		d.State = Available
		err = d.CommitFields(db, set.NewSet("state"))
		if err != nil {
			return
		}

		event.PublishDispatch(db, "disk.change")

		return
	}

	logrus.WithFields(logrus.Fields{
		"disk_id":   d.Id.Hex(),
		"disk_path": dskPath,
	}).Info("qemu: Destroying disk")

	err = utils.RemoveAll(dskPath)
	if err != nil {
		return
	}

	err = Remove(db, d.Id)
	if err != nil {
		return
	}

	return
}

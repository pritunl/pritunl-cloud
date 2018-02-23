package disk

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"gopkg.in/mgo.v2/bson"
	"strconv"
	"strings"
)

type Disk struct {
	Id           bson.ObjectId `bson:"_id,omitempty" json:"id"`
	Name         string        `bson:"name" json:"name"`
	State        string        `bson:"state" json:"state"`
	Node         bson.ObjectId `bson:"node" json:"node"`
	Organization bson.ObjectId `bson:"organization,omitempty" json:"organization"`
	Instance     bson.ObjectId `bson:"instance,omitempty" json:"instance"`
	Image        bson.ObjectId `bson:"image,omitempty" json:"image"`
	Index        string        `bson:"index" json:"index"`
	Size         int           `bson:"size" json:"size"`
}

func (d *Disk) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	if d.Instance != "" && d.Index != "" {
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

	if d.Instance == "" && !strings.HasPrefix(d.Index, "hold") {
		d.Index = fmt.Sprintf("hold_%s", bson.NewObjectId().Hex())
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

	err = coll.Insert(d)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func (d *Disk) Destroy(db *database.Database) (err error) {
	dskPath := vm.GetDiskPath(d.Id)

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

package instance

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/vm"
	"gopkg.in/mgo.v2/bson"
)

type Instance struct {
	Id           bson.ObjectId `bson:"_id,omitempty" json:"id"`
	Organization bson.ObjectId `bson:"organization,omitempty" json:"organization"`
	Zone         bson.ObjectId `bson:"zone,omitempty" json:"zone"`
	Status       string        `bson:"status" json:"status"`
	Node         bson.ObjectId `bson:"node,omitempty" json:"node"`
	Name         string        `bson:"name" json:"name"`
	Memory       int           `bson:"memory" json:"memory"`
	Processors   int           `bson:"processors" json:"processors"`
}

func (i *Instance) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	if i.Organization == "" {
		errData = &errortypes.ErrorData{
			Error:   "organization_required",
			Message: "Missing required organization",
		}
	}

	if i.Zone == "" {
		errData = &errortypes.ErrorData{
			Error:   "zone_required",
			Message: "Missing required zone",
		}
	}

	if i.Node == "" {
		errData = &errortypes.ErrorData{
			Error:   "node_required",
			Message: "Missing required node",
		}
	}

	if i.Memory < 256 {
		i.Memory = 256
	}

	if i.Processors < 1 {
		i.Processors = 1
	}

	return
}

func (i *Instance) Commit(db *database.Database) (err error) {
	coll := db.Instances()

	err = coll.Commit(i.Id, i)
	if err != nil {
		return
	}

	return
}

func (i *Instance) CommitFields(db *database.Database, fields set.Set) (
	err error) {

	coll := db.Instances()

	err = coll.CommitFields(i.Id, i, fields)
	if err != nil {
		return
	}

	return
}

func (i *Instance) Insert(db *database.Database) (err error) {
	coll := db.Instances()

	if i.Id != "" {
		err = &errortypes.DatabaseError{
			errors.New("datecenter: Instance already exists"),
		}
		return
	}

	err = coll.Insert(i)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func (i *Instance) GetVm() (virt *vm.VirtualMachine) {
	virt = &vm.VirtualMachine{
		Id:         i.Id,
		Processors: i.Processors,
		Memory:     i.Memory,
		Disks: []*vm.Disk{
			&vm.Disk{
				Path: vm.GetDiskPath(i.Id, 0),
			},
		},
		NetworkAdapters: []*vm.NetworkAdapter{
			&vm.NetworkAdapter{
				MacAddress:       vm.GetMacAddr(i.Id),
				BridgedInterface: vm.BridgedInterface,
			},
		},
	}

	return
}

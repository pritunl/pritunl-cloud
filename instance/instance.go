package instance

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/vm"
	"gopkg.in/mgo.v2/bson"
)

type Instance struct {
	Id           bson.ObjectId `bson:"_id,omitempty" json:"id"`
	Organization bson.ObjectId `bson:"organization,omitempty" json:"organization"`
	Zone         bson.ObjectId `bson:"zone,omitempty" json:"zone"`
	Status       string        `bson:"-" json:"status"`
	State        string        `bson:"state" json:"state"`
	VmState      string        `bson:"vm_state" json:"vm_state"`
	PublicIp     string        `bson:"public_ip" json:"public_ip"`
	PublicIp6    string        `bson:"public_ip6" json:"public_ip6"`
	Node         bson.ObjectId `bson:"node,omitempty" json:"node"`
	Name         string        `bson:"name" json:"name"`
	Memory       int           `bson:"memory" json:"memory"`
	Processors   int           `bson:"processors" json:"processors"`
}

func (i *Instance) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	if i.State == "" {
		i.State = Running
	}

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

func (i *Instance) Json() {
	switch i.State {
	case Running:
		switch i.VmState {
		case vm.Starting:
			i.Status = "Starting"
			break
		case vm.Running:
			i.Status = "Running"
			break
		case vm.Stopped:
			i.Status = "Starting"
			break
		case vm.Failed:
			i.Status = "Starting"
			break
		case vm.Updating:
			i.Status = "Updating"
			break
		case vm.ProvisioningDisk:
			i.Status = "Provisioning Disk"
			break
		case "":
			i.Status = "Provisioning"
			break
		}
		break
	case Stopped:
		switch i.VmState {
		case vm.Starting:
			i.Status = "Stopping"
			break
		case vm.Running:
			i.Status = "Stopping"
			break
		case vm.Stopped:
			i.Status = "Stopped"
			break
		case vm.Failed:
			i.Status = "Stopped"
			break
		case vm.Updating:
			i.Status = "Updating"
			break
		case vm.ProvisioningDisk:
			i.Status = "Provisioning Disk"
			break
		case "":
			i.Status = "Provisioning"
			break
		}
		break
	case Updating:
		i.Status = "Updating"
		break
	}
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
				BridgedInterface: node.Self.GetDefaultInterface(),
			},
		},
	}

	return
}

func (i *Instance) Changed(virt *vm.VirtualMachine) bool {
	return i.Memory != virt.Memory || i.Processors != virt.Processors
}

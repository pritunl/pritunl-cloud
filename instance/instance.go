package instance

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/bridge"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/pritunl/pritunl-cloud/vpc"
	"gopkg.in/mgo.v2/bson"
	"strconv"
)

type Instance struct {
	Id           bson.ObjectId      `bson:"_id,omitempty" json:"id"`
	Organization bson.ObjectId      `bson:"organization,omitempty" json:"organization"`
	Zone         bson.ObjectId      `bson:"zone,omitempty" json:"zone"`
	Vpc          bson.ObjectId      `bson:"vpc" json:"vpc"`
	Image        bson.ObjectId      `bson:"image,omitempty" json:"image"`
	Status       string             `bson:"-" json:"status"`
	State        string             `bson:"state" json:"state"`
	VmState      string             `bson:"vm_state" json:"vm_state"`
	Restart      bool               `bson:"restart" json:"restart"`
	PublicIps    []string           `bson:"public_ips" json:"public_ips"`
	PublicIps6   []string           `bson:"public_ips6" json:"public_ips6"`
	LocalIps     []string           `bson:"local_ips" json:"local_ips"`
	LocalIps6    []string           `bson:"local_ips6" json:"local_ips6"`
	Node         bson.ObjectId      `bson:"node,omitempty" json:"node"`
	Name         string             `bson:"name" json:"name"`
	Memory       int                `bson:"memory" json:"memory"`
	Processors   int                `bson:"processors" json:"processors"`
	NetworkRoles []string           `bson:"network_roles" json:"network_roles"`
	Virt         *vm.VirtualMachine `bson:"-" json:"-"`
	curVpc       bson.ObjectId      `bson:"-" json:"-"`
}

func (i *Instance) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	if i.State == "" {
		i.State = Start
	}

	if i.State != Start {
		i.Restart = false
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

	if i.Image == "" {
		errData = &errortypes.ErrorData{
			Error:   "image_required",
			Message: "Missing required image",
		}
	}

	if i.Vpc == "" {
		errData = &errortypes.ErrorData{
			Error:   "vpc_required",
			Message: "Missing required VPC",
		}
	}

	if i.Memory < 256 {
		i.Memory = 256
	}

	if i.Processors < 1 {
		i.Processors = 1
	}

	if i.NetworkRoles == nil {
		i.NetworkRoles = []string{}
	}

	if i.PublicIps == nil {
		i.PublicIps = []string{}
	}

	if i.PublicIps6 == nil {
		i.PublicIps6 = []string{}
	}

	if i.LocalIps == nil {
		i.LocalIps = []string{}
	}

	if i.LocalIps6 == nil {
		i.LocalIps6 = []string{}
	}

	return
}

func (i *Instance) Format() {
	// TODO Sort VPC IDs
}

func (i *Instance) Json() {
	switch i.State {
	case Start:
		if i.Restart {
			i.Status = "Restart Required"
		} else {
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
			case vm.Provisioning:
				i.Status = "Provisioning"
				break
			case "":
				i.Status = "Provisioning"
				break
			}
		}
		break
	case Stop:
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
		case vm.Provisioning:
			i.Status = "Provisioning"
			break
		case "":
			i.Status = "Provisioning"
			break
		}
		break
	case Restart:
		i.Status = "Restarting"
		break
	case Destroy:
		i.Status = "Destroying"
		break
	}
}

func (i *Instance) PreCommit() {
	i.curVpc = i.Vpc
}

func (i *Instance) PostCommit(db *database.Database) (err error) {
	if i.curVpc != "" && i.curVpc != i.Vpc {
		err = vpc.RemoveInstanceIp(db, i.Id, i.curVpc)
		if err != nil {
			return
		}
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
			errors.New("instance: Instance already exists"),
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

func (i *Instance) LoadVirt(disks []*disk.Disk) {
	i.Virt = &vm.VirtualMachine{
		Id:         i.Id,
		Image:      i.Image,
		Processors: i.Processors,
		Memory:     i.Memory,
		Disks:      []*vm.Disk{},
		NetworkAdapters: []*vm.NetworkAdapter{
			&vm.NetworkAdapter{
				Type:          vm.Bridge,
				MacAddress:    vm.GetMacAddr(i.Id, i.Vpc),
				HostInterface: bridge.BridgeName,
				VpcId:         i.Vpc,
			},
		},
	}

	if disks != nil {
		for _, dsk := range disks {
			index, err := strconv.Atoi(dsk.Index)
			if err != nil {
				continue
			}

			i.Virt.Disks = append(i.Virt.Disks, &vm.Disk{
				Index: index,
				Path:  paths.GetDiskPath(dsk.Id),
			})
		}
	}

	return
}

func (i *Instance) Changed(curVirt *vm.VirtualMachine) bool {
	if i.Virt.Memory != curVirt.Memory ||
		i.Virt.Processors != curVirt.Processors ||
		len(i.Virt.Disks) != len(curVirt.Disks) {

		return true
	}

	for i, dsk := range i.Virt.Disks {
		if len(curVirt.Disks) <= i {
			return true
		}

		if dsk.Index != curVirt.Disks[i].Index ||
			dsk.Path != curVirt.Disks[i].Path {

			return true
		}
	}

	for i, adapter := range i.Virt.NetworkAdapters {
		if len(curVirt.NetworkAdapters) <= i {
			return true
		}

		if adapter.VpcId != curVirt.NetworkAdapters[i].VpcId {
			return true
		}
	}

	return false
}

func (i *Instance) DiskChanged(curVirt *vm.VirtualMachine) (
	addDisks, remDisks []*vm.Disk) {

	addDisks = []*vm.Disk{}
	remDisks = []*vm.Disk{}
	disks := set.NewSet()
	curDisks := map[int]*vm.Disk{}

	for _, dsk := range i.Virt.Disks {
		disks.Add(dsk.Index)
	}

	for _, dsk := range curVirt.Disks {
		if !disks.Contains(dsk.Index) {
			remDisks = append(remDisks, dsk)
		} else {
			curDisks[dsk.Index] = dsk
		}
	}

	for _, dsk := range i.Virt.Disks {
		curDsk := curDisks[dsk.Index]
		if curDsk == nil {
			addDisks = append(addDisks, dsk)
		} else if dsk.Path != curDsk.Path {
			remDisks = append(remDisks, curDsk)
			addDisks = append(addDisks, dsk)
		}
	}

	return
}

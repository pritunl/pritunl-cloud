package state

import (
	"github.com/Sirupsen/logrus"
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/qemu"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/pritunl/pritunl-cloud/vpc"
	"gopkg.in/mgo.v2/bson"
)

type State struct {
	disks []*disk.Disk

	virtsMap     map[bson.ObjectId]*vm.VirtualMachine
	instances    []*instance.Instance
	vpcsMap      map[bson.ObjectId]*vpc.Vpc
	instancesMap map[bson.ObjectId]*instance.Instance
	addInstances set.Set
	remInstances set.Set
}

func (s *State) Instances() []*instance.Instance {
	return s.instances
}

func (s *State) Disks() []*disk.Disk {
	return s.disks
}

func (s *State) Vpc(vpcId bson.ObjectId) *vpc.Vpc {
	return s.vpcsMap[vpcId]
}

func (s *State) DiskInUse(instId, dskId bson.ObjectId) bool {
	curVirt := s.virtsMap[instId]

	if curVirt != nil {
		for _, vmDsk := range curVirt.Disks {
			if vmDsk.GetId() == dskId {
				return true
			}
		}
	}

	return false
}

func (s *State) GetVirt(instId bson.ObjectId) *vm.VirtualMachine {
	return s.virtsMap[instId]
}

func (s *State) init() (err error) {
	db := database.GetDatabase()
	defer db.Close()

	disks, err := disk.GetNode(db, node.Self.Id)
	if err != nil {
		return
	}
	s.disks = disks

	curVirts, err := qemu.GetVms(db)
	if err != nil {
		return
	}

	virtsId := set.NewSet()
	virtsMap := map[bson.ObjectId]*vm.VirtualMachine{}
	for _, virt := range curVirts {
		virtsId.Add(virt.Id)
		virtsMap[virt.Id] = virt
	}
	s.virtsMap = virtsMap

	instances, err := instance.GetAllVirt(db, &bson.M{
		"node": node.Self.Id,
	}, disks)
	s.instances = instances

	vpcIdsSet := set.NewSet()
	for _, inst := range instances {
		virtsId.Remove(inst.Id)
		vpcIdsSet.Add(inst.Vpc)
	}

	vpcIds := []bson.ObjectId{}
	for vpcIdInf := range vpcIdsSet.Iter() {
		vpcIds = append(vpcIds, vpcIdInf.(bson.ObjectId))
	}

	for virtId := range virtsId.Iter() {
		logrus.WithFields(logrus.Fields{
			"id": virtId.(bson.ObjectId).Hex(),
		}).Info("sync: Unknown instance")
	}
	s.instances = instances

	vpcs, err := vpc.GetIds(db, vpcIds)
	if err != nil {
		return
	}

	vpcsMap := map[bson.ObjectId]*vpc.Vpc{}
	for _, vc := range vpcs {
		vpcsMap[vc.Id] = vc
	}
	s.vpcsMap = vpcsMap

	return
}

func GetState() (stat *State, err error) {
	stat = &State{}

	err = stat.init()
	if err != nil {
		return
	}

	return
}

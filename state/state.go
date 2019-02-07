package state

import (
	"github.com/Sirupsen/logrus"
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/domain"
	"github.com/pritunl/pritunl-cloud/firewall"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/qemu"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/pritunl/pritunl-cloud/vpc"
	"github.com/pritunl/pritunl-cloud/zone"
)

type State struct {
	nodes            []*node.Node
	nodeZone         *zone.Zone
	namespaces       []string
	interfaces       []string
	nodeFirewall     []*firewall.Rule
	firewalls        map[string][]*firewall.Rule
	disks            []*disk.Disk
	virtsMap         map[primitive.ObjectID]*vm.VirtualMachine
	instances        []*instance.Instance
	instancesMap     map[primitive.ObjectID]*instance.Instance
	instanceDisks    map[primitive.ObjectID][]*disk.Disk
	domainRecordsMap map[primitive.ObjectID][]*domain.Record
	vpcsMap          map[primitive.ObjectID]*vpc.Vpc
	addInstances     set.Set
	remInstances     set.Set
}

func (s *State) Nodes() []*node.Node {
	return s.nodes
}

func (s *State) NodeZone() *zone.Zone {
	return s.nodeZone
}

func (s *State) Namespaces() []string {
	return s.namespaces
}

func (s *State) Interfaces() []string {
	return s.interfaces
}

func (s *State) Instances() []*instance.Instance {
	return s.instances
}

func (s *State) NodeFirewall() []*firewall.Rule {
	return s.nodeFirewall
}

func (s *State) Firewalls() map[string][]*firewall.Rule {
	return s.firewalls
}

func (s *State) DomainRecords(instId primitive.ObjectID) []*domain.Record {
	return s.domainRecordsMap[instId]
}

func (s *State) Disks() []*disk.Disk {
	return s.disks
}

func (s *State) GetInstaceDisks(instId primitive.ObjectID) []*disk.Disk {
	return s.instanceDisks[instId]
}

func (s *State) Vpc(vpcId primitive.ObjectID) *vpc.Vpc {
	return s.vpcsMap[vpcId]
}

func (s *State) DiskInUse(instId, dskId primitive.ObjectID) bool {
	curVirt := s.virtsMap[instId]

	if curVirt != nil {
		if curVirt.State != vm.Stopped && curVirt.State != vm.Failed {
			for _, vmDsk := range curVirt.Disks {
				if vmDsk.GetId() == dskId {
					return true
				}
			}
		}
	}

	return false
}

func (s *State) GetVirt(instId primitive.ObjectID) *vm.VirtualMachine {
	return s.virtsMap[instId]
}

func (s *State) GetInstace(instId primitive.ObjectID) *instance.Instance {
	return s.instancesMap[instId]
}

func (s *State) init() (err error) {
	db := database.GetDatabase()
	defer db.Close()

	zneId := node.Self.Zone
	if !zneId.IsZero() {
		zne, e := zone.Get(db, zneId)
		if e != nil {
			err = e
			return
		}

		s.nodeZone = zne
	}

	if s.nodeZone != nil && s.nodeZone.NetworkMode == zone.VxLan {
		ndes, e := node.GetAllNet(db)
		if e != nil {
			err = e
			return
		}

		s.nodes = ndes
	}

	namespaces, err := utils.GetNamespaces()
	if err != nil {
		return
	}
	s.namespaces = namespaces

	interfaces, err := utils.GetInterfaces()
	if err != nil {
		return
	}
	s.interfaces = interfaces

	disks, err := disk.GetNode(db, node.Self.Id)
	if err != nil {
		return
	}
	s.disks = disks

	instanceDisks := map[primitive.ObjectID][]*disk.Disk{}
	for _, dsk := range disks {
		dsks := instanceDisks[dsk.Instance]
		if dsks == nil {
			dsks = []*disk.Disk{}
		}
		instanceDisks[dsk.Instance] = append(dsks, dsk)
	}
	s.instanceDisks = instanceDisks

	curVirts, err := qemu.GetVms(db)
	if err != nil {
		return
	}

	virtsId := set.NewSet()
	virtsMap := map[primitive.ObjectID]*vm.VirtualMachine{}
	for _, virt := range curVirts {
		virtsId.Add(virt.Id)
		virtsMap[virt.Id] = virt
	}
	s.virtsMap = virtsMap

	instances, err := instance.GetAllVirtMapped(db, &bson.M{
		"node": node.Self.Id,
	}, instanceDisks)
	s.instances = instances

	instancesMap := map[primitive.ObjectID]*instance.Instance{}
	vpcIdsSet := set.NewSet()
	for _, inst := range instances {
		virtsId.Remove(inst.Id)
		vpcIdsSet.Add(inst.Vpc)
		instancesMap[inst.Id] = inst
	}
	s.instancesMap = instancesMap

	vpcIds := []primitive.ObjectID{}
	for vpcIdInf := range vpcIdsSet.Iter() {
		vpcIds = append(vpcIds, vpcIdInf.(primitive.ObjectID))
	}

	for virtId := range virtsId.Iter() {
		logrus.WithFields(logrus.Fields{
			"id": virtId.(primitive.ObjectID).Hex(),
		}).Info("sync: Unknown instance")
	}
	s.instances = instances

	nodeFirewall, firewalls, err := firewall.GetAllIngress(db, instances)
	if err != nil {
		return
	}
	s.nodeFirewall = nodeFirewall
	s.firewalls = firewalls

	vpcs, err := vpc.GetIds(db, vpcIds)
	if err != nil {
		return
	}

	vpcsMap := map[primitive.ObjectID]*vpc.Vpc{}
	for _, vc := range vpcs {
		vpcsMap[vc.Id] = vc
	}
	s.vpcsMap = vpcsMap

	recrds, err := domain.GetRecordAll(db, &bson.M{
		"node": node.Self.Id,
	})

	domainRecordsMap := map[primitive.ObjectID][]*domain.Record{}
	for _, recrd := range recrds {
		instRecrds := domainRecordsMap[recrd.Instance]
		if instRecrds == nil {
			instRecrds = []*domain.Record{}
		}

		instRecrds = append(instRecrds, recrd)
		domainRecordsMap[recrd.Instance] = instRecrds
	}
	s.domainRecordsMap = domainRecordsMap

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

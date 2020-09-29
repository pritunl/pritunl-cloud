package state

import (
	"io/ioutil"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/block"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/domain"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/firewall"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/qemu"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/pritunl/pritunl-cloud/vpc"
	"github.com/pritunl/pritunl-cloud/zone"
	"github.com/sirupsen/logrus"
)

type State struct {
	nodeSelf         *node.Node
	nodes            []*node.Node
	nodeDatacenter   primitive.ObjectID
	nodeZone         *zone.Zone
	nodeHostBlock    *block.Block
	vxlan            bool
	zoneMap          map[primitive.ObjectID]*zone.Zone
	namespaces       []string
	interfaces       []string
	interfacesSet    set.Set
	nodeFirewall     []*firewall.Rule
	firewalls        map[string][]*firewall.Rule
	disks            []*disk.Disk
	virtsMap         map[primitive.ObjectID]*vm.VirtualMachine
	instances        []*instance.Instance
	instancesMap     map[primitive.ObjectID]*instance.Instance
	instanceDisks    map[primitive.ObjectID][]*disk.Disk
	domainRecordsMap map[primitive.ObjectID][]*domain.Record
	vpcs             []*vpc.Vpc
	vpcsMap          map[primitive.ObjectID]*vpc.Vpc
	addInstances     set.Set
	remInstances     set.Set
	running          []string
}

func (s *State) Node() *node.Node {
	return s.nodeSelf
}

func (s *State) Nodes() []*node.Node {
	return s.nodes
}

func (s *State) VxLan() bool {
	return s.vxlan
}

func (s *State) NodeZone() *zone.Zone {
	return s.nodeZone
}

func (s *State) NodeHostBlock() *block.Block {
	return s.nodeHostBlock
}

func (s *State) GetZone(zneId primitive.ObjectID) *zone.Zone {
	return s.zoneMap[zneId]
}

func (s *State) Namespaces() []string {
	return s.namespaces
}

func (s *State) Interfaces() []string {
	return s.interfaces
}

func (s *State) HasInterfaces(iface string) bool {
	return s.interfacesSet.Contains(iface)
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

func (s *State) Running() []string {
	return s.running
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

func (s *State) Vpcs() []*vpc.Vpc {
	return s.vpcs
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

	s.nodeSelf = node.Self.Copy()

	zneId := s.nodeSelf.Zone
	if !zneId.IsZero() {
		zne, e := zone.Get(db, zneId)
		if e != nil {
			err = e
			return
		}

		s.nodeZone = zne
		s.nodeDatacenter = s.nodeZone.Datacenter
	}

	if s.nodeZone != nil && s.nodeZone.NetworkMode == zone.VxlanVlan {
		s.vxlan = true

		znes, e := zone.GetAllDatacenter(db, s.nodeZone.Datacenter)
		if e != nil {
			err = e
			return
		}

		zonesMap := map[primitive.ObjectID]*zone.Zone{}
		for _, zne := range znes {
			zonesMap[zne.Id] = zne
		}
		s.zoneMap = zonesMap

		ndes, e := node.GetAllNet(db)
		if e != nil {
			err = e
			return
		}

		s.nodes = ndes
	}

	hostBlockId := s.nodeSelf.HostBlock
	if !hostBlockId.IsZero() {
		hostBlock, e := block.Get(db, hostBlockId)
		if e != nil {
			err = e
			if _, ok := err.(*database.NotFoundError); ok {
				hostBlock = nil
				err = nil
			} else {
				return
			}
		}

		s.nodeHostBlock = hostBlock
	}

	namespaces, err := utils.GetNamespaces()
	if err != nil {
		return
	}
	s.namespaces = namespaces

	interfaces, interfacesSet, err := utils.GetInterfacesSet()
	if err != nil {
		return
	}
	s.interfaces = interfaces
	s.interfacesSet = interfacesSet

	disks, err := disk.GetNode(db, s.nodeSelf.Id)
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

	instances, err := instance.GetAllVirtMapped(db, &bson.M{
		"node": s.nodeSelf.Id,
	}, instanceDisks)
	s.instances = instances

	instId := set.NewSet()
	instancesMap := map[primitive.ObjectID]*instance.Instance{}
	for _, inst := range instances {
		instId.Add(inst.Id)
		instancesMap[inst.Id] = inst
	}
	s.instancesMap = instancesMap

	curVirts, err := qemu.GetVms(db, instancesMap)
	if err != nil {
		return
	}

	virtsMap := map[primitive.ObjectID]*vm.VirtualMachine{}
	for _, virt := range curVirts {
		if !instId.Contains(virt.Id) {
			logrus.WithFields(logrus.Fields{
				"id": virt.Id.Hex(),
			}).Info("sync: Unknown instance")
		}
		virtsMap[virt.Id] = virt
	}
	s.virtsMap = virtsMap

	nodeFirewall, firewalls, err := firewall.GetAllIngress(
		db, s.nodeSelf, instances)
	if err != nil {
		return
	}
	s.nodeFirewall = nodeFirewall
	s.firewalls = firewalls

	vpcs := []*vpc.Vpc{}
	vpcsMap := map[primitive.ObjectID]*vpc.Vpc{}
	if !s.nodeDatacenter.IsZero() {
		vpcs, err = vpc.GetDatacenter(db, s.nodeDatacenter)
		if err != nil {
			return
		}

		for _, vc := range vpcs {
			vpcsMap[vc.Id] = vc
		}
	}
	s.vpcs = vpcs
	s.vpcsMap = vpcsMap

	recrds, err := domain.GetRecordAll(db, &bson.M{
		"node": s.nodeSelf.Id,
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

	items, err := ioutil.ReadDir("/var/run")
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "state: Failed to read run directory"),
		}
		return
	}

	running := []string{}
	for _, item := range items {
		if !item.IsDir() {
			running = append(running, item.Name())
		}
	}
	s.running = running

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

package state

import (
	"io/ioutil"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/arp"
	"github.com/pritunl/pritunl-cloud/block"
	"github.com/pritunl/pritunl-cloud/certificate"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/deployment"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/firewall"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/pool"
	"github.com/pritunl/pritunl-cloud/qemu"
	"github.com/pritunl/pritunl-cloud/scheduler"
	"github.com/pritunl/pritunl-cloud/secret"
	"github.com/pritunl/pritunl-cloud/service"
	"github.com/pritunl/pritunl-cloud/shape"
	"github.com/pritunl/pritunl-cloud/spec"
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
	nodeShapes       []*shape.Shape
	nodeShapesId     set.Set
	vxlan            bool
	zoneMap          map[primitive.ObjectID]*zone.Zone
	namespaces       []string
	interfaces       []string
	interfacesSet    set.Set
	nodeFirewall     []*firewall.Rule
	firewalls        map[string][]*firewall.Rule
	pools            []*pool.Pool
	disks            []*disk.Disk
	schedulers       []*scheduler.Scheduler
	deploymentsMap   map[primitive.ObjectID]*deployment.Deployment
	servicesMap      map[primitive.ObjectID]*service.Service
	servicesUnitsMap map[primitive.ObjectID]*service.Unit

	specsMap            map[spec.Hash]*spec.Spec
	specsServicesMap    map[primitive.ObjectID]*service.Service
	specsDeploymentsMap map[primitive.ObjectID]*deployment.Deployment
	specsSecretsMap     map[primitive.ObjectID]*secret.Secret
	specsCertsMap       map[primitive.ObjectID]*certificate.Certificate

	virtsMap      map[primitive.ObjectID]*vm.VirtualMachine
	instances     []*instance.Instance
	instancesMap  map[primitive.ObjectID]*instance.Instance
	instanceDisks map[primitive.ObjectID][]*disk.Disk
	vpcs          []*vpc.Vpc
	vpcsMap       map[primitive.ObjectID]*vpc.Vpc
	vpcIpsMap     map[primitive.ObjectID][]*vpc.VpcIp
	arpRecords    map[string]set.Set
	addInstances  set.Set
	remInstances  set.Set
	running       []string
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

func (s *State) Schedulers() []*scheduler.Scheduler {
	return s.schedulers
}

func (s *State) NodeFirewall() []*firewall.Rule {
	return s.nodeFirewall
}

func (s *State) Firewalls() map[string][]*firewall.Rule {
	return s.firewalls
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

func (s *State) Deployment(deplyId primitive.ObjectID) *deployment.Deployment {
	return s.deploymentsMap[deplyId]
}

func (s *State) Deployments() (
	deplys map[primitive.ObjectID]*deployment.Deployment) {

	deplys = s.deploymentsMap
	return
}

func (s *State) Service(srvcId primitive.ObjectID) *service.Service {
	return s.servicesMap[srvcId]
}

func (s *State) Unit(unitId primitive.ObjectID) *service.Unit {
	return s.servicesUnitsMap[unitId]
}

func (s *State) Spec(hash spec.Hash) *spec.Spec {
	return s.specsMap[hash]
}

func (s *State) SpecService(srvcId primitive.ObjectID) *service.Service {
	return s.specsServicesMap[srvcId]
}

func (s *State) SpecSecret(secrID primitive.ObjectID) *secret.Secret {
	return s.specsSecretsMap[secrID]
}

func (s *State) SpecCert(certId primitive.ObjectID) *certificate.Certificate {
	return s.specsCertsMap[certId]
}

func (s *State) Vpc(vpcId primitive.ObjectID) *vpc.Vpc {
	return s.vpcsMap[vpcId]
}

func (s *State) VpcIps(vpcId primitive.ObjectID) []*vpc.VpcIp {
	return s.vpcIpsMap[vpcId]
}

func (s *State) VpcIpsMap() map[primitive.ObjectID][]*vpc.VpcIp {
	return s.vpcIpsMap
}

func (s *State) ArpRecords(namespace string) set.Set {
	return s.arpRecords[namespace]
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
	if instId.IsZero() {
		return nil
	}
	return s.virtsMap[instId]
}

func (s *State) GetInstace(instId primitive.ObjectID) *instance.Instance {
	if instId.IsZero() {
		return nil
	}
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

	pools, err := pool.GetAll(db, &bson.M{
		"zone": s.nodeSelf.Zone,
	})
	if err != nil {
		return
	}
	s.pools = pools

	disks, err := disk.GetNode(db, s.nodeSelf.Id, s.nodeSelf.Pools)
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
	}, s.pools, instanceDisks)
	if err != nil {
		return
	}

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

	shapes, err := shape.GetAll(db, &bson.M{
		"roles": &bson.M{
			"$in": node.Self.NetworkRoles,
		},
	})
	if err != nil {
		return
	}
	s.nodeShapes = shapes

	nodeShapesId := set.NewSet()
	for _, shape := range shapes {
		nodeShapesId.Add(shape.Id)
	}
	s.nodeShapesId = nodeShapesId

	vpcs := []*vpc.Vpc{}
	vpcsId := []primitive.ObjectID{}
	vpcsMap := map[primitive.ObjectID]*vpc.Vpc{}
	if !s.nodeDatacenter.IsZero() {
		vpcs, err = vpc.GetDatacenter(db, s.nodeDatacenter)
		if err != nil {
			return
		}

		for _, vc := range vpcs {
			vpcsId = append(vpcsId, vc.Id)
			vpcsMap[vc.Id] = vc
		}
	}
	s.vpcs = vpcs
	s.vpcsMap = vpcsMap

	vpcIpsMap := map[primitive.ObjectID][]*vpc.VpcIp{}
	if !s.nodeDatacenter.IsZero() {
		vpcIpsMap, err = vpc.GetIpsMapped(db, vpcsId)
		if err != nil {
			return
		}
	}
	s.vpcIpsMap = vpcIpsMap

	s.arpRecords = arp.BuildState(s.instances, s.vpcIpsMap)

	deployments, err := deployment.GetAll(db, &bson.M{
		"node": node.Self.Id,
	})
	if err != nil {
		return
	}

	deploymentsMap := map[primitive.ObjectID]*deployment.Deployment{}
	serviceIds := []primitive.ObjectID{}
	serviceIdsSet := set.NewSet()
	unitIds := set.NewSet()
	specIdsSet := set.NewSet()
	for _, deply := range deployments {
		deploymentsMap[deply.Id] = deply
		serviceIdsSet.Add(deply.Service)
		unitIds.Add(deply.Unit)
		specIdsSet.Add(deply.GetSpecHash())
	}

	specIds := []spec.Hash{}
	for specId := range specIdsSet.Iter() {
		specIds = append(specIds, specId.(spec.Hash))
	}

	specs, err := spec.GetAll(db, &bson.M{
		"_id": &bson.M{
			"$in": specIds,
		},
	})
	if err != nil {
		return
	}

	specSecretsSet := set.NewSet()
	specCertsSet := set.NewSet()
	specServicesSet := set.NewSet()
	specsMap := map[spec.Hash]*spec.Spec{}
	for _, spc := range specs {
		specsMap[spc.Id] = spc

		if spc.Instance.Services != nil {
			for _, srvcId := range spc.Instance.Services {
				specServicesSet.Add(srvcId)
			}
		}

		if spc.Instance.Secrets != nil {
			for _, secrId := range spc.Instance.Secrets {
				specSecretsSet.Add(secrId)
			}
		}

		if spc.Instance.Certificates != nil {
			for _, certId := range spc.Instance.Certificates {
				specCertsSet.Add(certId)
			}
		}
	}
	s.specsMap = specsMap

	specServiceIds := []primitive.ObjectID{}
	for srvcId := range specServicesSet.Iter() {
		specServiceIds = append(specServiceIds, srvcId.(primitive.ObjectID))
	}

	specDeploymentsSet := set.NewSet()
	specsServicesMap := map[primitive.ObjectID]*service.Service{}
	specServices, err := service.GetAll(db, &bson.M{
		"_id": &bson.M{
			"$in": specServiceIds,
		},
	})
	if err != nil {
		return
	}

	for _, specService := range specServices {
		specsServicesMap[specService.Id] = specService

		for _, unit := range specService.Units {
			for _, deply := range unit.Deployments {
				specDeploymentsSet.Add(deply.Id)
			}
		}
	}
	s.specsServicesMap = specsServicesMap

	specDeploymentIds := []primitive.ObjectID{}
	for deplyIdInf := range specDeploymentsSet.Iter() {
		deplyId := deplyIdInf.(primitive.ObjectID)
		if _, exists := deploymentsMap[deplyId]; !exists {
			specDeploymentIds = append(specDeploymentIds, deplyId)
		}
	}

	specDeployments, err := deployment.GetAll(db, &bson.M{
		"_id": &bson.M{
			"$in": specDeploymentIds,
		},
	})
	if err != nil {
		return
	}

	for _, specDeployment := range specDeployments {
		deploymentsMap[specDeployment.Id] = specDeployment
	}
	s.deploymentsMap = deploymentsMap

	for serviceId := range serviceIdsSet.Iter() {
		serviceIds = append(serviceIds, serviceId.(primitive.ObjectID))
	}

	services, err := service.GetAll(db, &bson.M{
		"_id": &bson.M{
			"$in": serviceIds,
		},
	})
	if err != nil {
		return
	}

	serviceDeploymentsSet := set.NewSet()
	servicesMap := map[primitive.ObjectID]*service.Service{}
	servicesUnitsMap := map[primitive.ObjectID]*service.Unit{}
	for _, srvc := range services {
		servicesMap[srvc.Id] = srvc

		for _, unit := range srvc.Units {
			if !unitIds.Contains(unit.Id) ||
				unit.Kind != spec.InstanceKind {

				continue
			}
			servicesUnitsMap[unit.Id] = unit

			for _, deply := range unit.Deployments {
				serviceDeploymentsSet.Add(deply.Id)
			}
		}
	}
	s.servicesMap = servicesMap
	s.servicesUnitsMap = servicesUnitsMap

	serviceDeploymentIds := []primitive.ObjectID{}
	for deplyIdInf := range serviceDeploymentsSet.Iter() {
		deplyId := deplyIdInf.(primitive.ObjectID)
		if _, exists := deploymentsMap[deplyId]; !exists {
			serviceDeploymentIds = append(serviceDeploymentIds, deplyId)
		}
	}

	serviceDeployments, err := deployment.GetAll(db, &bson.M{
		"_id": &bson.M{
			"$in": serviceDeploymentIds,
		},
	})
	if err != nil {
		return
	}

	for _, serviceDeployment := range serviceDeployments {
		deploymentsMap[serviceDeployment.Id] = serviceDeployment
	}
	s.deploymentsMap = deploymentsMap

	schedulers, err := scheduler.GetAll(db)
	if err != nil {
		return
	}
	s.schedulers = schedulers

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

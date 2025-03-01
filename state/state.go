package state

import (
	"io/ioutil"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/arp"
	"github.com/pritunl/pritunl-cloud/certificate"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/datacenter"
	"github.com/pritunl/pritunl-cloud/deployment"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/domain"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/firewall"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/nodeport"
	"github.com/pritunl/pritunl-cloud/pod"
	"github.com/pritunl/pritunl-cloud/pool"
	"github.com/pritunl/pritunl-cloud/qemu"
	"github.com/pritunl/pritunl-cloud/scheduler"
	"github.com/pritunl/pritunl-cloud/secret"
	"github.com/pritunl/pritunl-cloud/shape"
	"github.com/pritunl/pritunl-cloud/spec"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/pritunl/pritunl-cloud/vpc"
	"github.com/pritunl/pritunl-cloud/zone"
	"github.com/sirupsen/logrus"
)

type State struct {
	nodeSelf               *node.Node
	nodes                  []*node.Node
	nodeDatacenter         *datacenter.Datacenter
	nodeZone               *zone.Zone
	nodeShapes             []*shape.Shape
	nodeShapesId           set.Set
	vxlan                  bool
	datacenterMap          map[primitive.ObjectID]*datacenter.Datacenter
	zoneMap                map[primitive.ObjectID]*zone.Zone
	namespaces             []string
	interfaces             []string
	interfacesSet          set.Set
	nodeFirewall           []*firewall.Rule
	firewalls              map[string][]*firewall.Rule
	firewallMaps           map[string][]*firewall.Mapping
	pools                  []*pool.Pool
	disks                  []*disk.Disk
	schedulers             []*scheduler.Scheduler
	deploymentsReservedMap map[primitive.ObjectID]*deployment.Deployment
	deploymentsDeployedMap map[primitive.ObjectID]*deployment.Deployment
	deploymentsInactiveMap map[primitive.ObjectID]*deployment.Deployment
	podsMap                map[primitive.ObjectID]*pod.Pod
	podsUnitsMap           map[primitive.ObjectID]*pod.Unit

	specsMap            map[primitive.ObjectID]*spec.Spec
	specsPodsMap        map[primitive.ObjectID]*pod.Pod
	specsPodsUnitsMap   map[primitive.ObjectID]*pod.Unit
	specsDeploymentsMap map[primitive.ObjectID]*deployment.Deployment
	specsDomainsMap     map[primitive.ObjectID]*domain.Domain
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

func (s *State) NodeDatacenter() *datacenter.Datacenter {
	return s.nodeDatacenter
}

func (s *State) NodeZone() *zone.Zone {
	return s.nodeZone
}

func (s *State) GetDatacenter(dcId primitive.ObjectID) *datacenter.Datacenter {
	return s.datacenterMap[dcId]
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

func (s *State) FirewallMaps() map[string][]*firewall.Mapping {
	return s.firewallMaps
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

func (s *State) DeploymentReserved(deplyId primitive.ObjectID) *deployment.Deployment {
	return s.deploymentsReservedMap[deplyId]
}

func (s *State) DeploymentsReserved() (
	deplys map[primitive.ObjectID]*deployment.Deployment) {

	deplys = s.deploymentsReservedMap
	return
}

func (s *State) DeploymentDeployed(deplyId primitive.ObjectID) *deployment.Deployment {
	return s.deploymentsDeployedMap[deplyId]
}

func (s *State) DeploymentsDeployed() (
	deplys map[primitive.ObjectID]*deployment.Deployment) {

	deplys = s.deploymentsDeployedMap
	return
}

func (s *State) DeploymentsDestroy() (
	deplys map[primitive.ObjectID]*deployment.Deployment) {

	deplys = s.deploymentsInactiveMap
	return
}

func (s *State) DeploymentInactive(deplyId primitive.ObjectID) *deployment.Deployment {
	return s.deploymentsInactiveMap[deplyId]
}

func (s *State) DeploymentsInactive() (
	deplys map[primitive.ObjectID]*deployment.Deployment) {

	deplys = s.deploymentsInactiveMap
	return
}

func (s *State) Deployment(deplyId primitive.ObjectID) (
	deply *deployment.Deployment) {

	deply = s.deploymentsDeployedMap[deplyId]
	if deply != nil {
		return
	}

	deply = s.deploymentsReservedMap[deplyId]
	if deply != nil {
		return
	}

	deply = s.deploymentsInactiveMap[deplyId]
	if deply != nil {
		return
	}

	return
}

func (s *State) Pod(pdId primitive.ObjectID) *pod.Pod {
	return s.podsMap[pdId]
}

func (s *State) Unit(unitId primitive.ObjectID) *pod.Unit {
	return s.podsUnitsMap[unitId]
}

func (s *State) Spec(commitId primitive.ObjectID) *spec.Spec {
	return s.specsMap[commitId]
}

func (s *State) SpecPod(pdId primitive.ObjectID) *pod.Pod {
	return s.specsPodsMap[pdId]
}

func (s *State) SpecUnit(unitId primitive.ObjectID) *pod.Unit {
	return s.specsPodsUnitsMap[unitId]
}

func (s *State) SpecDomain(domnId primitive.ObjectID) *domain.Domain {
	return s.specsDomainsMap[domnId]
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

func (s *State) init(runtimes *Runtimes) (err error) {
	db := database.GetDatabase()
	defer db.Close()

	start := time.Now()
	totalStart := time.Now()
	s.nodeSelf = node.Self.Copy()

	dcId := s.nodeSelf.Datacenter
	if !dcId.IsZero() {
		dc, e := datacenter.Get(db, dcId)
		if e != nil {
			err = e
			return
		}

		s.nodeDatacenter = dc
	}

	zneId := s.nodeSelf.Zone
	if !zneId.IsZero() {
		zne, e := zone.Get(db, zneId)
		if e != nil {
			err = e
			return
		}

		s.nodeZone = zne
	}

	if s.nodeDatacenter != nil &&
		s.nodeDatacenter.NetworkMode == datacenter.VxlanVlan {

		s.vxlan = true

		znes, e := zone.GetAllDatacenter(db, s.nodeDatacenter.Id)
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

	runtimes.State1 = time.Since(start)
	start = time.Now()

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

	runtimes.State2 = time.Since(start)
	start = time.Now()

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

	runtimes.State3 = time.Since(start)
	start = time.Now()

	shapes := []*shape.Shape{}
	nodeNetworkRoles := node.Self.NetworkRoles
	if len(nodeNetworkRoles) > 0 {
		shapes, err = shape.GetAll(db, &bson.M{
			"roles": &bson.M{
				"$in": nodeNetworkRoles,
			},
		})
		if err != nil {
			return
		}
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
	if s.nodeDatacenter != nil {
		vpcs, err = vpc.GetDatacenter(db, s.nodeDatacenter.Id)
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
	if s.nodeDatacenter != nil {
		vpcIpsMap, err = vpc.GetIpsMapped(db, vpcsId)
		if err != nil {
			return
		}
	}
	s.vpcIpsMap = vpcIpsMap

	runtimes.State4 = time.Since(start)
	start = time.Now()

	deployments, err := deployment.GetAll(db, &bson.M{
		"node": node.Self.Id,
	})
	if err != nil {
		return
	}

	deploymentsNode := map[primitive.ObjectID]*deployment.Deployment{}
	deploymentsReservedMap := map[primitive.ObjectID]*deployment.Deployment{}
	deploymentsDeployedMap := map[primitive.ObjectID]*deployment.Deployment{}
	deploymentsInactiveMap := map[primitive.ObjectID]*deployment.Deployment{}
	deploymentsIdSet := set.NewSet()
	podIds := []primitive.ObjectID{}
	podIdsSet := set.NewSet()
	unitIds := set.NewSet()
	specIdsSet := set.NewSet()
	for _, deply := range deployments {
		deploymentsNode[deply.Id] = deply

		deploymentsIdSet.Add(deply.Id)
		switch deply.State {
		case deployment.Reserved:
			deploymentsReservedMap[deply.Id] = deply
			break
		case deployment.Deployed, deployment.Migrate:
			deploymentsDeployedMap[deply.Id] = deply
			break
		case deployment.Destroy, deployment.Archive,
			deployment.Archived, deployment.Restore:

			deploymentsInactiveMap[deply.Id] = deply
			break
		}

		podIdsSet.Add(deply.Pod)
		unitIds.Add(deply.Unit)
		specIdsSet.Add(deply.Spec)
	}

	runtimes.State5 = time.Since(start)
	start = time.Now()

	specIds := []primitive.ObjectID{}
	for specId := range specIdsSet.Iter() {
		specIds = append(specIds, specId.(primitive.ObjectID))
	}

	specs := []*spec.Spec{}
	if len(specIds) > 0 {
		specs, err = spec.GetAll(db, &bson.M{
			"_id": &bson.M{
				"$in": specIds,
			},
		})
		if err != nil {
			return
		}
	}

	specSecretsSet := set.NewSet()
	specCertsSet := set.NewSet()
	specPodsSet := set.NewSet()
	specDomainsSet := set.NewSet()
	specsMap := map[primitive.ObjectID]*spec.Spec{}
	for _, spc := range specs {
		specsMap[spc.Id] = spc

		if spc.Instance != nil {
			if spc.Instance.Pods != nil {
				for _, pdId := range spc.Instance.Pods {
					specPodsSet.Add(pdId)
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

		if spc.Firewall != nil {
			for _, rule := range spc.Firewall.Ingress {
				for _, ref := range rule.Sources {
					specPodsSet.Add(ref.Realm)
				}
			}
		}

		if spc.Domain != nil {
			for _, record := range spc.Domain.Records {
				specDomainsSet.Add(record.Domain)
			}
		}
	}
	s.specsMap = specsMap

	specCertIds := []primitive.ObjectID{}
	for certId := range specCertsSet.Iter() {
		specCertIds = append(specCertIds, certId.(primitive.ObjectID))
	}

	specsCertsMap := map[primitive.ObjectID]*certificate.Certificate{}
	specCerts := []*certificate.Certificate{}
	if len(specCertIds) > 0 {
		specCerts, err = certificate.GetAll(db, &bson.M{
			"_id": &bson.M{
				"$in": specCertIds,
			},
		})
		if err != nil {
			return
		}
	}

	for _, specCert := range specCerts {
		specsCertsMap[specCert.Id] = specCert
	}
	s.specsCertsMap = specsCertsMap

	specSecretIds := []primitive.ObjectID{}
	for secrId := range specSecretsSet.Iter() {
		specSecretIds = append(specSecretIds, secrId.(primitive.ObjectID))
	}

	runtimes.State6 = time.Since(start)
	start = time.Now()

	specsSecretsMap := map[primitive.ObjectID]*secret.Secret{}

	specSecrets := []*secret.Secret{}
	if len(specSecretIds) > 0 {
		specSecrets, err = secret.GetAll(db, &bson.M{
			"_id": &bson.M{
				"$in": specSecretIds,
			},
		})
		if err != nil {
			return
		}
	}

	for _, specSecret := range specSecrets {
		specsSecretsMap[specSecret.Id] = specSecret
	}
	s.specsSecretsMap = specsSecretsMap

	specPodIds := []primitive.ObjectID{}
	for pdId := range specPodsSet.Iter() {
		specPodIds = append(specPodIds, pdId.(primitive.ObjectID))
	}

	specPods := []*pod.Pod{}
	if len(specPodIds) > 0 {
		specPods, err = pod.GetAll(db, &bson.M{
			"_id": &bson.M{
				"$in": specPodIds,
			},
		})
		if err != nil {
			return
		}
	}

	specDeploymentsSet := set.NewSet()
	specsPodsMap := map[primitive.ObjectID]*pod.Pod{}
	specsPodsUnitsMap := map[primitive.ObjectID]*pod.Unit{}
	for _, specPod := range specPods {
		specsPodsMap[specPod.Id] = specPod

		for _, unit := range specPod.Units {
			specsPodsUnitsMap[unit.Id] = unit
			for _, deply := range unit.Deployments {
				specDeploymentsSet.Add(deply.Id)
			}
		}
	}
	s.specsPodsMap = specsPodsMap
	s.specsPodsUnitsMap = specsPodsUnitsMap

	specDomainIds := []primitive.ObjectID{}
	for pdId := range specDomainsSet.Iter() {
		specDomainIds = append(specDomainIds, pdId.(primitive.ObjectID))
	}

	specsDomainsMap := map[primitive.ObjectID]*domain.Domain{}
	specDomains, err := domain.GetLoadedAllIds(db, specDomainIds)
	if err != nil {
		return
	}

	for _, specDomain := range specDomains {
		specsDomainsMap[specDomain.Id] = specDomain
	}
	s.specsDomainsMap = specsDomainsMap

	specDeploymentIds := []primitive.ObjectID{}
	for deplyIdInf := range specDeploymentsSet.Iter() {
		deplyId := deplyIdInf.(primitive.ObjectID)
		if !deploymentsIdSet.Contains(deplyId) {
			specDeploymentIds = append(specDeploymentIds, deplyId)
		}
	}

	specDeployments := []*deployment.Deployment{}
	if len(specDeploymentIds) > 0 {
		specDeployments, err = deployment.GetAll(db, &bson.M{
			"_id": &bson.M{
				"$in": specDeploymentIds,
			},
		})
		if err != nil {
			return
		}
	}

	for _, specDeployment := range specDeployments {
		deploymentsIdSet.Add(specDeployment.Id)
		switch specDeployment.State {
		case deployment.Reserved:
			deploymentsReservedMap[specDeployment.Id] = specDeployment
			break
		case deployment.Deployed, deployment.Migrate:
			deploymentsDeployedMap[specDeployment.Id] = specDeployment
			break
		case deployment.Destroy, deployment.Archive,
			deployment.Archived, deployment.Restore:

			deploymentsInactiveMap[specDeployment.Id] = specDeployment
			break
		}
	}

	runtimes.State7 = time.Since(start)
	start = time.Now()

	for podId := range podIdsSet.Iter() {
		podIds = append(podIds, podId.(primitive.ObjectID))
	}

	pods := []*pod.Pod{}
	if len(podIds) > 0 {
		pods, err = pod.GetAll(db, &bson.M{
			"_id": &bson.M{
				"$in": podIds,
			},
		})
		if err != nil {
			return
		}
	}

	nodePortsDeployments := map[primitive.ObjectID][]*nodeport.Mapping{}
	podDeploymentsSet := set.NewSet()
	podsMap := map[primitive.ObjectID]*pod.Pod{}
	podsUnitsMap := map[primitive.ObjectID]*pod.Unit{}
	for _, pd := range pods {
		podsMap[pd.Id] = pd

		for _, unit := range pd.Units {
			if !unitIds.Contains(unit.Id) {
				continue
			}
			podsUnitsMap[unit.Id] = unit

			for _, deply := range unit.Deployments {
				nodePortsDeployments[deply.Id] = append(
					nodePortsDeployments[deply.Id], unit.NodePorts...)

				podDeploymentsSet.Add(deply.Id)
			}
		}
	}
	s.podsMap = podsMap
	s.podsUnitsMap = podsUnitsMap

	podDeploymentIds := []primitive.ObjectID{}
	for deplyIdInf := range podDeploymentsSet.Iter() {
		deplyId := deplyIdInf.(primitive.ObjectID)
		if !deploymentsIdSet.Contains(deplyId) {
			podDeploymentIds = append(podDeploymentIds, deplyId)
		}
	}

	podDeployments := []*deployment.Deployment{}
	if len(podDeploymentIds) > 0 {
		podDeployments, err = deployment.GetAll(db, &bson.M{
			"_id": &bson.M{
				"$in": podDeploymentIds,
			},
		})
		if err != nil {
			return
		}
	}

	for _, podDeployment := range podDeployments {
		deploymentsIdSet.Add(podDeployment.Id)
		switch podDeployment.State {
		case deployment.Reserved:
			deploymentsReservedMap[podDeployment.Id] = podDeployment
			break
		case deployment.Deployed, deployment.Migrate:
			deploymentsDeployedMap[podDeployment.Id] = podDeployment
			break
		case deployment.Destroy, deployment.Archive,
			deployment.Archived, deployment.Restore:

			deploymentsInactiveMap[podDeployment.Id] = podDeployment
			break
		}
	}
	s.deploymentsReservedMap = deploymentsReservedMap
	s.deploymentsDeployedMap = deploymentsDeployedMap
	s.deploymentsInactiveMap = deploymentsInactiveMap

	runtimes.State8 = time.Since(start)
	start = time.Now()

	instances, err := instance.GetAllVirtMapped(db, &bson.M{
		"node": s.nodeSelf.Id,
	}, s.pools, instanceDisks)
	if err != nil {
		return
	}

	s.instances = instances

	nodePortsMap := map[string][]*nodeport.Mapping{}

	instId := set.NewSet()
	instancesMap := map[primitive.ObjectID]*instance.Instance{}
	for _, inst := range instances {
		instId.Add(inst.Id)
		instancesMap[inst.Id] = inst

		nodePortsMap[inst.NetworkNamespace] = append(
			nodePortsMap[inst.NetworkNamespace], inst.NodePorts...)

		if !inst.Deployment.IsZero() {
			nodePortsMap[inst.NetworkNamespace] = append(
				nodePortsMap[inst.NetworkNamespace],
				nodePortsDeployments[inst.Deployment]...,
			)
		}
	}
	s.instancesMap = instancesMap

	runtimes.State9 = time.Since(start)
	start = time.Now()

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

	s.arpRecords = arp.BuildState(s.instances, s.vpcsMap, s.vpcIpsMap)

	runtimes.State10 = time.Since(start)
	start = time.Now()

	specRules, err := firewall.GetSpecRules(instances, deploymentsNode,
		specsMap, specsPodsUnitsMap, deploymentsDeployedMap)
	if err != nil {
		return
	}

	nodeFirewall, firewalls, firewallMaps, err := firewall.GetAllIngress(
		db, s.nodeSelf, instances, specRules, nodePortsMap)
	if err != nil {
		return
	}
	s.nodeFirewall = nodeFirewall
	s.firewalls = firewalls
	s.firewallMaps = firewallMaps

	schedulers, err := scheduler.GetAll(db)
	if err != nil {
		return
	}
	s.schedulers = schedulers

	runtimes.State11 = time.Since(start)
	start = time.Now()

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

	runtimes.State12 = time.Since(start)
	runtimes.State = time.Since(totalStart)

	return
}

func GetState(runtimes *Runtimes) (stat *State, err error) {
	stat = &State{}

	err = stat.init(runtimes)
	if err != nil {
		return
	}

	return
}

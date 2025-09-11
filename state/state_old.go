package state

import (
	"io/ioutil"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/arp"
	"github.com/pritunl/pritunl-cloud/authority"
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
	"github.com/pritunl/pritunl-cloud/spec"
	"github.com/pritunl/pritunl-cloud/unit"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/pritunl/pritunl-cloud/vpc"
	"github.com/pritunl/pritunl-cloud/zone"
	"github.com/sirupsen/logrus"
)

type StateOld struct {
	nodeSelf               *node.Node
	nodes                  []*node.Node
	nodeDatacenter         *datacenter.Datacenter
	nodeZone               *zone.Zone
	vxlan                  bool
	zoneMap                map[bson.ObjectID]*zone.Zone
	namespaces             []string
	interfaces             []string
	interfacesSet          set.Set
	nodeFirewall           []*firewall.Rule
	firewalls              map[string][]*firewall.Rule
	firewallMaps           map[string][]*firewall.Mapping
	pools                  []*pool.Pool
	disks                  []*disk.Disk
	schedulers             []*scheduler.Scheduler
	deploymentsReservedMap map[bson.ObjectID]*deployment.Deployment
	deploymentsDeployedMap map[bson.ObjectID]*deployment.Deployment
	deploymentsInactiveMap map[bson.ObjectID]*deployment.Deployment
	podsMap                map[bson.ObjectID]*pod.Pod
	unitsMap               map[bson.ObjectID]*unit.Unit

	specsMap            map[bson.ObjectID]*spec.Spec
	specsPodsMap        map[bson.ObjectID]*pod.Pod
	specsPodUnitsMap    map[bson.ObjectID][]*unit.Unit
	specsUnitsMap       map[bson.ObjectID]*unit.Unit
	specsDeploymentsMap map[bson.ObjectID]*deployment.Deployment
	specsDomainsMap     map[bson.ObjectID]*domain.Domain
	specsSecretsMap     map[bson.ObjectID]*secret.Secret
	specsCertsMap       map[bson.ObjectID]*certificate.Certificate

	virtsMap           map[bson.ObjectID]*vm.VirtualMachine
	instances          []*instance.Instance
	instancesMap       map[bson.ObjectID]*instance.Instance
	instanceDisks      map[bson.ObjectID][]*disk.Disk
	instanceNamespaces map[bson.ObjectID][]string
	authoritiesMap     map[string][]*authority.Authority
	vpcs               []*vpc.Vpc
	vpcsMap            map[bson.ObjectID]*vpc.Vpc
	vpcIpsMap          map[bson.ObjectID][]*vpc.VpcIp
	arpRecords         map[string]set.Set
	addInstances       set.Set
	remInstances       set.Set
	running            []string
}

func (s *StateOld) Node() *node.Node {
	return s.nodeSelf
}

func (s *StateOld) Nodes() []*node.Node {
	return s.nodes
}

func (s *StateOld) VxLan() bool {
	return s.vxlan
}

func (s *StateOld) NodeDatacenter() *datacenter.Datacenter {
	return s.nodeDatacenter
}

func (s *StateOld) NodeZone() *zone.Zone {
	return s.nodeZone
}

func (s *StateOld) GetZone(zneId bson.ObjectID) *zone.Zone {
	return s.zoneMap[zneId]
}

func (s *StateOld) Namespaces() []string {
	return s.namespaces
}

func (s *StateOld) Interfaces() []string {
	return s.interfaces
}

func (s *StateOld) HasInterfaces(iface string) bool {
	return s.interfacesSet.Contains(iface)
}

func (s *StateOld) Instances() []*instance.Instance {
	return s.instances
}

func (s *StateOld) Schedulers() []*scheduler.Scheduler {
	return s.schedulers
}

func (s *StateOld) NodeFirewall() []*firewall.Rule {
	return s.nodeFirewall
}

func (s *StateOld) Firewalls() map[string][]*firewall.Rule {
	return s.firewalls
}

func (s *StateOld) FirewallMaps() map[string][]*firewall.Mapping {
	return s.firewallMaps
}

func (s *StateOld) Running() []string {
	return s.running
}

func (s *StateOld) Disks() []*disk.Disk {
	return s.disks
}

func (s *StateOld) GetInstaceDisks(instId bson.ObjectID) []*disk.Disk {
	return s.instanceDisks[instId]
}

func (s *StateOld) GetInstanceNamespaces(instId bson.ObjectID) []string {
	return s.instanceNamespaces[instId]
}

func (s *StateOld) GetInstaceAuthorities(roles []string) []*authority.Authority {
	authrSet := set.NewSet()
	authrs := []*authority.Authority{}

	for _, role := range roles {
		for _, authr := range s.authoritiesMap[role] {
			if authrSet.Contains(authr.Id) {
				continue
			}
			authrSet.Add(authr.Id)
			authrs = append(authrs, authr)
		}
	}

	return authrs
}

func (s *StateOld) DeploymentReserved(deplyId bson.ObjectID) *deployment.Deployment {
	return s.deploymentsReservedMap[deplyId]
}

func (s *StateOld) DeploymentsReserved() (
	deplys map[bson.ObjectID]*deployment.Deployment) {

	deplys = s.deploymentsReservedMap
	return
}

func (s *StateOld) DeploymentDeployed(deplyId bson.ObjectID) *deployment.Deployment {
	return s.deploymentsDeployedMap[deplyId]
}

func (s *StateOld) DeploymentsDeployed() (
	deplys map[bson.ObjectID]*deployment.Deployment) {

	deplys = s.deploymentsDeployedMap
	return
}

func (s *StateOld) DeploymentsDestroy() (
	deplys map[bson.ObjectID]*deployment.Deployment) {

	deplys = s.deploymentsInactiveMap
	return
}

func (s *StateOld) DeploymentInactive(deplyId bson.ObjectID) *deployment.Deployment {
	return s.deploymentsInactiveMap[deplyId]
}

func (s *StateOld) DeploymentsInactive() (
	deplys map[bson.ObjectID]*deployment.Deployment) {

	deplys = s.deploymentsInactiveMap
	return
}

func (s *StateOld) Deployment(deplyId bson.ObjectID) (
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

func (s *StateOld) Pod(pdId bson.ObjectID) *pod.Pod {
	return s.podsMap[pdId]
}

func (s *StateOld) Unit(unitId bson.ObjectID) *unit.Unit {
	return s.unitsMap[unitId]
}

func (s *StateOld) Spec(commitId bson.ObjectID) *spec.Spec {
	return s.specsMap[commitId]
}

func (s *StateOld) SpecPod(pdId bson.ObjectID) *pod.Pod {
	return s.specsPodsMap[pdId]
}

func (s *StateOld) SpecPodUnits(pdId bson.ObjectID) []*unit.Unit {
	return s.specsPodUnitsMap[pdId]
}

func (s *StateOld) SpecUnit(unitId bson.ObjectID) *unit.Unit {
	return s.specsUnitsMap[unitId]
}

func (s *StateOld) SpecDomain(domnId bson.ObjectID) *domain.Domain {
	return s.specsDomainsMap[domnId]
}

func (s *StateOld) SpecSecret(secrID bson.ObjectID) *secret.Secret {
	return s.specsSecretsMap[secrID]
}

func (s *StateOld) SpecCert(certId bson.ObjectID) *certificate.Certificate {
	return s.specsCertsMap[certId]
}

func (s *StateOld) Vpc(vpcId bson.ObjectID) *vpc.Vpc {
	return s.vpcsMap[vpcId]
}

func (s *StateOld) VpcIps(vpcId bson.ObjectID) []*vpc.VpcIp {
	return s.vpcIpsMap[vpcId]
}

func (s *StateOld) VpcIpsMap() map[bson.ObjectID][]*vpc.VpcIp {
	return s.vpcIpsMap
}

func (s *StateOld) ArpRecords(namespace string) set.Set {
	return s.arpRecords[namespace]
}

func (s *StateOld) Vpcs() []*vpc.Vpc {
	return s.vpcs
}

func (s *StateOld) DiskInUse(instId, dskId bson.ObjectID) bool {
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

func (s *StateOld) GetVirt(instId bson.ObjectID) *vm.VirtualMachine {
	if instId.IsZero() {
		return nil
	}
	return s.virtsMap[instId]
}

func (s *StateOld) GetInstace(instId bson.ObjectID) *instance.Instance {
	if instId.IsZero() {
		return nil
	}
	return s.instancesMap[instId]
}

func (s *StateOld) init() (err error) {
	db := database.GetDatabase()
	defer db.Close()

	s.nodeSelf = node.Self.Copy()

	// Datacenter
	dcId := s.nodeSelf.Datacenter
	if !dcId.IsZero() {
		dc, e := datacenter.Get(db, dcId)
		if e != nil {
			err = e
			return
		}

		s.nodeDatacenter = dc
	}
	// Datacenter

	// Zone
	zneId := s.nodeSelf.Zone
	if !zneId.IsZero() {
		zne, e := zone.Get(db, zneId)
		if e != nil {
			err = e
			return
		}

		s.nodeZone = zne
	}
	// Zone

	// Zones
	if s.nodeDatacenter != nil && s.nodeDatacenter.Vxlan() {
		s.vxlan = true

		znes, e := zone.GetAllDatacenter(db, s.nodeDatacenter.Id)
		if e != nil {
			err = e
			return
		}

		zonesMap := map[bson.ObjectID]*zone.Zone{}
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
	// Zones

	// Network
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
	// Network

	// Pools
	pools, err := pool.GetAll(db, &bson.M{
		"zone": s.nodeSelf.Zone,
	})
	if err != nil {
		return
	}
	s.pools = pools
	// Pools

	// Disks
	disks, err := disk.GetNode(db, s.nodeSelf.Id, s.nodeSelf.Pools)
	if err != nil {
		return
	}
	s.disks = disks

	instanceDisks := map[bson.ObjectID][]*disk.Disk{}
	for _, dsk := range disks {
		dsks := instanceDisks[dsk.Instance]
		if dsks == nil {
			dsks = []*disk.Disk{}
		}
		instanceDisks[dsk.Instance] = append(dsks, dsk)
	}
	s.instanceDisks = instanceDisks
	// Disks

	// Vpcs
	vpcs := []*vpc.Vpc{}
	vpcsId := []bson.ObjectID{}
	vpcsMap := map[bson.ObjectID]*vpc.Vpc{}
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

	vpcIpsMap := map[bson.ObjectID][]*vpc.VpcIp{}
	if s.nodeDatacenter != nil {
		vpcIpsMap, err = vpc.GetIpsMapped(db, vpcsId)
		if err != nil {
			return
		}
	}
	s.vpcIpsMap = vpcIpsMap
	// Vpcs

	// Deployments
	deployments, err := deployment.GetAll(db, &bson.M{
		"node": node.Self.Id,
	})
	if err != nil {
		return
	}

	deploymentsNode := map[bson.ObjectID]*deployment.Deployment{}
	deploymentsReservedMap := map[bson.ObjectID]*deployment.Deployment{}
	deploymentsDeployedMap := map[bson.ObjectID]*deployment.Deployment{}
	deploymentsInactiveMap := map[bson.ObjectID]*deployment.Deployment{}
	deploymentsIdSet := set.NewSet()
	podIdsSet := set.NewSet()
	unitIdsSet := set.NewSet()
	specIdsSet := set.NewSet()
	for _, deply := range deployments {
		deploymentsNode[deply.Id] = deply

		deploymentsIdSet.Add(deply.Id)
		switch deply.State {
		case deployment.Reserved:
			deploymentsReservedMap[deply.Id] = deply
			break
		case deployment.Deployed:
			switch deply.Action {
			case deployment.Destroy, deployment.Archive,
				deployment.Restore:

				deploymentsInactiveMap[deply.Id] = deply
				break
			default:
				deploymentsDeployedMap[deply.Id] = deply
			}
			break
		case deployment.Archived:
			deploymentsInactiveMap[deply.Id] = deply
			break
		}

		podIdsSet.Add(deply.Pod)
		unitIdsSet.Add(deply.Unit)
		specIdsSet.Add(deply.Spec)
	}

	specIds := []bson.ObjectID{}
	for specId := range specIdsSet.Iter() {
		specIds = append(specIds, specId.(bson.ObjectID))
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
	specUnitsSet := set.NewSet()
	specDomainsSet := set.NewSet()
	specsMap := map[bson.ObjectID]*spec.Spec{}
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
					specUnitsSet.Add(ref.Id)
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

	specCertIds := []bson.ObjectID{}
	for certId := range specCertsSet.Iter() {
		specCertIds = append(specCertIds, certId.(bson.ObjectID))
	}

	specsCertsMap := map[bson.ObjectID]*certificate.Certificate{}
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

	specSecretIds := []bson.ObjectID{}
	for secrId := range specSecretsSet.Iter() {
		specSecretIds = append(specSecretIds, secrId.(bson.ObjectID))
	}

	specsSecretsMap := map[bson.ObjectID]*secret.Secret{}

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

	specPodIds := []bson.ObjectID{}
	for pdId := range specPodsSet.Iter() {
		specPodIds = append(specPodIds, pdId.(bson.ObjectID))
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

	specsPodsMap := map[bson.ObjectID]*pod.Pod{}
	for _, specPod := range specPods {
		specsPodsMap[specPod.Id] = specPod
	}
	s.specsPodsMap = specsPodsMap

	specUnitIds := []bson.ObjectID{}
	for unitId := range specUnitsSet.Iter() {
		specUnitIds = append(specUnitIds, unitId.(bson.ObjectID))
	}

	specUnits := []*unit.Unit{}
	if len(specUnitIds) > 0 || len(specPodIds) > 0 {
		specUnits, err = unit.GetAll(db, &bson.M{
			"$or": []*bson.M{
				&bson.M{
					"_id": &bson.M{
						"$in": specUnitIds,
					},
				},
				&bson.M{
					"pod": &bson.M{
						"$in": specPodIds,
					},
				},
			},
		})
		if err != nil {
			return
		}
	}

	specDeploymentsSet := set.NewSet()
	specsUnitsMap := map[bson.ObjectID]*unit.Unit{}
	specsPodUnitsMap := map[bson.ObjectID][]*unit.Unit{}
	for _, specUnit := range specUnits {
		specsUnitsMap[specUnit.Id] = specUnit

		specsPodUnitsMap[specUnit.Pod] = append(
			specsPodUnitsMap[specUnit.Pod], specUnit)

		for _, deplyId := range specUnit.Deployments {
			specDeploymentsSet.Add(deplyId)
		}
	}
	s.specsUnitsMap = specsUnitsMap
	s.specsPodUnitsMap = specsPodUnitsMap

	specDomainIds := []bson.ObjectID{}
	for pdId := range specDomainsSet.Iter() {
		specDomainIds = append(specDomainIds, pdId.(bson.ObjectID))
	}

	specsDomainsMap := map[bson.ObjectID]*domain.Domain{}
	specDomains, err := domain.GetLoadedAllIds(db, specDomainIds)
	if err != nil {
		return
	}

	for _, specDomain := range specDomains {
		specsDomainsMap[specDomain.Id] = specDomain
	}
	s.specsDomainsMap = specsDomainsMap

	specDeploymentIds := []bson.ObjectID{}
	for deplyIdInf := range specDeploymentsSet.Iter() {
		deplyId := deplyIdInf.(bson.ObjectID)
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
		case deployment.Deployed:
			switch specDeployment.Action {
			case deployment.Destroy, deployment.Archive,
				deployment.Restore:

				deploymentsInactiveMap[specDeployment.Id] = specDeployment
				break
			default:
				deploymentsDeployedMap[specDeployment.Id] = specDeployment
			}
			break
		case deployment.Archived:
			deploymentsInactiveMap[specDeployment.Id] = specDeployment
			break
		}
	}

	podIds := []bson.ObjectID{}
	for podId := range podIdsSet.Iter() {
		podIds = append(podIds, podId.(bson.ObjectID))
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

	podsMap := map[bson.ObjectID]*pod.Pod{}
	for _, pd := range pods {
		podsMap[pd.Id] = pd
	}
	s.podsMap = podsMap

	unitIds := []bson.ObjectID{}
	for unitId := range unitIdsSet.Iter() {
		unitIds = append(unitIds, unitId.(bson.ObjectID))
	}

	units := []*unit.Unit{}
	if len(unitIds) > 0 {
		units, err = unit.GetAll(db, &bson.M{
			"_id": &bson.M{
				"$in": unitIds,
			},
		})
		if err != nil {
			return
		}
	}

	unitsMap := map[bson.ObjectID]*unit.Unit{}
	podDeploymentsSet := set.NewSet()
	for _, unt := range units {
		unitsMap[unt.Id] = unt

		for _, deplyId := range unt.Deployments {
			podDeploymentsSet.Add(deplyId)
		}

	}
	s.unitsMap = unitsMap

	podDeploymentIds := []bson.ObjectID{}
	for deplyIdInf := range podDeploymentsSet.Iter() {
		deplyId := deplyIdInf.(bson.ObjectID)
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
		case deployment.Deployed:
			switch podDeployment.Action {
			case deployment.Destroy, deployment.Archive,
				deployment.Restore:

				deploymentsInactiveMap[podDeployment.Id] = podDeployment
				break
			default:
				deploymentsDeployedMap[podDeployment.Id] = podDeployment
			}
			break
		case deployment.Archived:
			deploymentsInactiveMap[podDeployment.Id] = podDeployment
			break
		}
	}
	s.deploymentsReservedMap = deploymentsReservedMap
	s.deploymentsDeployedMap = deploymentsDeployedMap
	s.deploymentsInactiveMap = deploymentsInactiveMap
	// Deployments

	// Instances
	instances, err := instance.GetAllVirtMapped(db, &bson.M{
		"node": s.nodeSelf.Id,
	}, s.pools, instanceDisks)
	if err != nil {
		return
	}

	s.instances = instances

	nodePortsMap := map[string][]*nodeport.Mapping{}

	instId := set.NewSet()
	instancesMap := map[bson.ObjectID]*instance.Instance{}
	instancesRolesSet := set.NewSet()
	for _, inst := range instances {
		instId.Add(inst.Id)
		instancesMap[inst.Id] = inst

		nodePortsMap[inst.NetworkNamespace] = append(
			nodePortsMap[inst.NetworkNamespace], inst.NodePorts...)

		for _, role := range inst.Roles {
			instancesRolesSet.Add(role)
		}
	}
	s.instancesMap = instancesMap

	instancesRoles := []string{}
	for instRoleInf := range instancesRolesSet.Iter() {
		instancesRoles = append(instancesRoles, instRoleInf.(string))
	}

	authrsMap, err := authority.GetMapRoles(db, &bson.M{
		"roles": &bson.M{
			"$in": instancesRoles,
		},
	})
	if err != nil {
		return
	}
	s.authoritiesMap = authrsMap
	// Instances

	// Virtuals
	curVirts, err := qemu.GetVms(db)
	if err != nil {
		return
	}

	virtsMap := map[bson.ObjectID]*vm.VirtualMachine{}
	for _, virt := range curVirts {
		if !instId.Contains(virt.Id) {
			logrus.WithFields(logrus.Fields{
				"id": virt.Id.Hex(),
			}).Info("sync: Unknown instance")
		}
		virtsMap[virt.Id] = virt
	}
	s.virtsMap = virtsMap
	// Virtuals

	// Firewalls
	s.arpRecords = arp.BuildState(s.instances, s.vpcsMap, s.vpcIpsMap)
	// Firewalls

	// Firewalls
	specRules, err := firewall.GetSpecRules(instances, deploymentsNode,
		specsMap, specsUnitsMap, deploymentsDeployedMap)
	if err != nil {
		return
	}

	nodeFirewall, firewalls, firewallMaps, instNamespaces, err :=
		firewall.GetAllIngress(db, s.nodeSelf, instances,
			specRules, nodePortsMap)
	if err != nil {
		return
	}
	s.nodeFirewall = nodeFirewall
	s.firewalls = firewalls
	s.firewallMaps = firewallMaps
	s.instanceNamespaces = instNamespaces
	// Firewalls

	// Schedulers
	schedulers, err := scheduler.GetAll(db)
	if err != nil {
		return
	}
	s.schedulers = schedulers
	// Schedulers

	// Running
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
	// Running

	return
}

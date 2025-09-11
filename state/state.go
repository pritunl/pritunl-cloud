package state

import (
	"sync"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/authority"
	"github.com/pritunl/pritunl-cloud/certificate"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/datacenter"
	"github.com/pritunl/pritunl-cloud/deployment"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/domain"
	"github.com/pritunl/pritunl-cloud/firewall"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/nodeport"
	"github.com/pritunl/pritunl-cloud/pod"
	"github.com/pritunl/pritunl-cloud/pool"
	"github.com/pritunl/pritunl-cloud/scheduler"
	"github.com/pritunl/pritunl-cloud/secret"
	"github.com/pritunl/pritunl-cloud/spec"
	"github.com/pritunl/pritunl-cloud/unit"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/pritunl/pritunl-cloud/vpc"
	"github.com/pritunl/pritunl-cloud/zone"
)

type State struct {
	waiter *sync.WaitGroup

	// Datacenter
	NodeDatacenter func() *datacenter.Datacenter

	// Zone
	NodeZone func() *zone.Zone

	// Zones
	VxLan   func() bool
	GetZone func(zneId bson.ObjectID) *zone.Zone
	Nodes   func() []*node.Node

	// Network
	Namespaces    func() []string
	Interfaces    func() []string
	HasInterfaces func(iface string) bool

	// Pools
	NodePools func() []*pool.Pool

	// Disks
	Disks              func() []*disk.Disk
	GetInstaceDisks    func(instId bson.ObjectID) []*disk.Disk
	GetDeploymentDisks func(deplyId bson.ObjectID) []*disk.Disk
	InstaceDisksMap    func() map[bson.ObjectID][]*disk.Disk

	// Vpcs
	Vpc       func(vpcId bson.ObjectID) *vpc.Vpc
	VpcsMap   func() map[bson.ObjectID]*vpc.Vpc
	VpcIps    func(vpcId bson.ObjectID) []*vpc.VpcIp
	VpcIpsMap func() map[bson.ObjectID][]*vpc.VpcIp
	Vpcs      func() []*vpc.Vpc

	// Deployments
	Pod                 func(pdId bson.ObjectID) *pod.Pod
	PodsMap             func() map[bson.ObjectID]*pod.Pod
	Unit                func(unitId bson.ObjectID) *unit.Unit
	UnitsMap            func() map[bson.ObjectID]*unit.Unit
	Spec                func(commitId bson.ObjectID) *spec.Spec
	SpecsMap            func() map[bson.ObjectID]*spec.Spec
	SpecPod             func(pdId bson.ObjectID) *pod.Pod
	SpecPodUnits        func(pdId bson.ObjectID) []*unit.Unit
	SpecUnit            func(unitId bson.ObjectID) *unit.Unit
	SpecsUnitsMap       func() map[bson.ObjectID]*unit.Unit
	SpecDomain          func(domnId bson.ObjectID) *domain.Domain
	SpecSecret          func(secrID bson.ObjectID) *secret.Secret
	SpecCert            func(certId bson.ObjectID) *certificate.Certificate
	DeploymentsNode     func() map[bson.ObjectID]*deployment.Deployment
	DeploymentReserved  func(deplyId bson.ObjectID) *deployment.Deployment
	DeploymentsReserved func() map[bson.ObjectID]*deployment.Deployment
	DeploymentDeployed  func(deplyId bson.ObjectID) *deployment.Deployment
	DeploymentsDeployed func() map[bson.ObjectID]*deployment.Deployment
	DeploymentsDestroy  func() map[bson.ObjectID]*deployment.Deployment
	DeploymentInactive  func(deplyId bson.ObjectID) *deployment.Deployment
	DeploymentsInactive func() map[bson.ObjectID]*deployment.Deployment
	Deployment          func(deplyId bson.ObjectID) *deployment.Deployment

	// Instances
	GetInstace            func(instId bson.ObjectID) *instance.Instance
	Instances             func() []*instance.Instance
	NodePortsMap          func() map[string][]*nodeport.Mapping
	GetInstaceAuthorities func(orgId bson.ObjectID,
		roles []string) []*authority.Authority

	// Virtuals
	DiskInUse func(instId, dskId bson.ObjectID) bool
	GetVirt   func(instId bson.ObjectID) *vm.VirtualMachine
	VirtsMap  func() map[bson.ObjectID]*vm.VirtualMachine

	// Schedulers
	Schedulers func() []*scheduler.Scheduler

	// Firewalls
	NodeFirewall          func() []*firewall.Rule
	Firewalls             func() map[string][]*firewall.Rule
	FirewallMaps          func() map[string][]*firewall.Mapping
	ArpRecords            func(namespace string) set.Set
	GetInstanceNamespaces func(instId bson.ObjectID) []string
}

func (s *State) Node() *node.Node {
	return node.Self
}

func (s *State) WaitAdd() {
	s.waiter.Add(1)
}

func (s *State) WaitDone() {
	s.waiter.Done()
}

func (s *State) Wait() {
	s.waiter.Wait()
}

func GetState(runtimes *Runtimes) (stat *State, err error) {
	db := database.GetDatabase()
	defer db.Close()

	err = RefreshAll(db, runtimes)
	if err != nil {
		return
	}

	stat = &State{
		waiter: &sync.WaitGroup{},
	}
	ApplyAll(stat)

	return
}

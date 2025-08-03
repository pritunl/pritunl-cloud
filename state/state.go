package state

import (
	"sync"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
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
	GetZone func(zneId primitive.ObjectID) *zone.Zone
	Nodes   func() []*node.Node

	// Network
	Namespaces    func() []string
	Interfaces    func() []string
	HasInterfaces func(iface string) bool

	// Pools
	NodePools func() []*pool.Pool

	// Disks
	Disks              func() []*disk.Disk
	GetInstaceDisks    func(instId primitive.ObjectID) []*disk.Disk
	GetDeploymentDisks func(deplyId primitive.ObjectID) []*disk.Disk
	InstaceDisksMap    func() map[primitive.ObjectID][]*disk.Disk

	// Vpcs
	Vpc       func(vpcId primitive.ObjectID) *vpc.Vpc
	VpcsMap   func() map[primitive.ObjectID]*vpc.Vpc
	VpcIps    func(vpcId primitive.ObjectID) []*vpc.VpcIp
	VpcIpsMap func() map[primitive.ObjectID][]*vpc.VpcIp
	Vpcs      func() []*vpc.Vpc

	// Deployments
	Pod                 func(pdId primitive.ObjectID) *pod.Pod
	PodsMap             func() map[primitive.ObjectID]*pod.Pod
	Unit                func(unitId primitive.ObjectID) *unit.Unit
	UnitsMap            func() map[primitive.ObjectID]*unit.Unit
	Spec                func(commitId primitive.ObjectID) *spec.Spec
	SpecsMap            func() map[primitive.ObjectID]*spec.Spec
	SpecPod             func(pdId primitive.ObjectID) *pod.Pod
	SpecPodUnits        func(pdId primitive.ObjectID) []*unit.Unit
	SpecUnit            func(unitId primitive.ObjectID) *unit.Unit
	SpecsUnitsMap       func() map[primitive.ObjectID]*unit.Unit
	SpecDomain          func(domnId primitive.ObjectID) *domain.Domain
	SpecSecret          func(secrID primitive.ObjectID) *secret.Secret
	SpecCert            func(certId primitive.ObjectID) *certificate.Certificate
	DeploymentsNode     func() map[primitive.ObjectID]*deployment.Deployment
	DeploymentReserved  func(deplyId primitive.ObjectID) *deployment.Deployment
	DeploymentsReserved func() map[primitive.ObjectID]*deployment.Deployment
	DeploymentDeployed  func(deplyId primitive.ObjectID) *deployment.Deployment
	DeploymentsDeployed func() map[primitive.ObjectID]*deployment.Deployment
	DeploymentsDestroy  func() map[primitive.ObjectID]*deployment.Deployment
	DeploymentInactive  func(deplyId primitive.ObjectID) *deployment.Deployment
	DeploymentsInactive func() map[primitive.ObjectID]*deployment.Deployment
	Deployment          func(deplyId primitive.ObjectID) *deployment.Deployment

	// Instances
	GetInstace            func(instId primitive.ObjectID) *instance.Instance
	Instances             func() []*instance.Instance
	NodePortsMap          func() map[string][]*nodeport.Mapping
	GetInstaceAuthorities func(orgId primitive.ObjectID,
		roles []string) []*authority.Authority

	// Virtuals
	DiskInUse func(instId, dskId primitive.ObjectID) bool
	GetVirt   func(instId primitive.ObjectID) *vm.VirtualMachine
	VirtsMap  func() map[primitive.ObjectID]*vm.VirtualMachine

	// Schedulers
	Schedulers func() []*scheduler.Scheduler

	// Firewalls
	NodeFirewall          func() []*firewall.Rule
	Firewalls             func() map[string][]*firewall.Rule
	FirewallMaps          func() map[string][]*firewall.Mapping
	ArpRecords            func(namespace string) set.Set
	GetInstanceNamespaces func(instId primitive.ObjectID) []string
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

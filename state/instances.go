package state

import (
	"sync"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/nodeport"
)

var (
	Instances    = &InstancesState{}
	InstancesPkg = NewPackage(Instances)
)

type InstancesState struct {
	roles        []string
	rolesSet     set.Set
	rolesLock    sync.Mutex
	instances    []*instance.Instance
	instancesMap map[primitive.ObjectID]*instance.Instance
	nodePortsMap map[string][]*nodeport.Mapping
}

func (p *InstancesState) GetRoles() (roles []string, rolesSet set.Set) {
	p.rolesLock.Lock()
	roles = p.roles
	rolesSet = p.rolesSet
	p.rolesLock.Unlock()
	return
}

func (p *InstancesState) setRoles(roles []string, rolesSet set.Set) {
	p.rolesLock.Lock()
	p.roles = roles
	p.rolesSet = rolesSet
	p.rolesLock.Unlock()
	return
}

func (p *InstancesState) GetInstace(
	instId primitive.ObjectID) *instance.Instance {

	if instId.IsZero() {
		return nil
	}
	return p.instancesMap[instId]
}

func (p *InstancesState) Instances() []*instance.Instance {
	return p.instances
}

func (p *InstancesState) NodePortsMap() map[string][]*nodeport.Mapping {
	return p.nodePortsMap
}

func (p *InstancesState) Refresh(pkg *Package,
	db *database.Database) (err error) {

	ndeId := node.Self.Id

	instances, err := instance.GetAllVirtMapped(db, &bson.M{
		"node": ndeId,
	}, Pools.NodePools(), Disks.InstaceDisksMap())
	if err != nil {
		return
	}

	p.instances = instances

	instId := set.NewSet()
	instancesMap := map[primitive.ObjectID]*instance.Instance{}
	rolesSet := set.NewSet()
	nodePortsMap := map[string][]*nodeport.Mapping{}
	for _, inst := range instances {
		instId.Add(inst.Id)
		instancesMap[inst.Id] = inst

		nodePortsMap[inst.NetworkNamespace] = append(
			nodePortsMap[inst.NetworkNamespace], inst.NodePorts...)

		for _, role := range inst.NetworkRoles {
			rolesSet.Add(role)
		}
	}
	p.instancesMap = instancesMap
	p.nodePortsMap = nodePortsMap

	nde := node.Self
	if nde.Firewall {
		roles := nde.NetworkRoles
		for _, role := range roles {
			rolesSet.Add(role)
		}
	}

	roles := []string{}
	for instRoleInf := range rolesSet.Iter() {
		roles = append(roles, instRoleInf.(string))
	}

	p.setRoles(roles, rolesSet)

	return
}

func (p *InstancesState) Apply(st *State) {
	st.GetInstace = p.GetInstace
	st.Instances = p.Instances
	st.NodePortsMap = p.NodePortsMap
}

func init() {
	InstancesPkg.
		After(Disks).
		After(Pools)
}

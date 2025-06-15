package state

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/authority"
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
	instances      []*instance.Instance
	instancesMap   map[primitive.ObjectID]*instance.Instance
	authoritiesMap map[string][]*authority.Authority
	nodePortsMap   map[string][]*nodeport.Mapping
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

func (p *InstancesState) GetInstaceAuthorities(
	orgId primitive.ObjectID, roles []string) []*authority.Authority {

	authrSet := set.NewSet()
	authrs := []*authority.Authority{}

	for _, role := range roles {
		for _, authr := range p.authoritiesMap[role] {
			if authrSet.Contains(authr.Id) || authr.Organization != orgId {
				continue
			}
			authrSet.Add(authr.Id)
			authrs = append(authrs, authr)
		}
	}

	return authrs
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
	instancesRolesSet := set.NewSet()
	nodePortsMap := map[string][]*nodeport.Mapping{}
	for _, inst := range instances {
		instId.Add(inst.Id)
		instancesMap[inst.Id] = inst

		nodePortsMap[inst.NetworkNamespace] = append(
			nodePortsMap[inst.NetworkNamespace], inst.NodePorts...)

		for _, role := range inst.NetworkRoles {
			instancesRolesSet.Add(role)
		}
	}
	p.instancesMap = instancesMap
	p.nodePortsMap = nodePortsMap

	instancesRoles := []string{}
	for instRoleInf := range instancesRolesSet.Iter() {
		instancesRoles = append(instancesRoles, instRoleInf.(string))
	}

	authrsMap, err := authority.GetMapRoles(db, &bson.M{
		"network_roles": &bson.M{
			"$in": instancesRoles,
		},
	})
	if err != nil {
		return
	}
	p.authoritiesMap = authrsMap

	return
}

func (p *InstancesState) Apply(st *State) {
	st.GetInstace = p.GetInstace
	st.Instances = p.Instances
	st.NodePortsMap = p.NodePortsMap
	st.GetInstaceAuthorities = p.GetInstaceAuthorities
}

func init() {
	InstancesPkg.
		After(Disks).
		After(Pools)
}

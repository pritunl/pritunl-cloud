package state

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/nodeport"
)

var (
	Instances    = &InstancesState{}
	InstancesPkg = NewPackage(Instances)
)

type InstancesState struct {
	instances    []*instance.Instance
	instancesMap map[primitive.ObjectID]*instance.Instance
	nodePortsMap map[string][]*nodeport.Mapping
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

	instances := InstancesPreload.GetInstances()
	instances = instance.LoadAllVirt(instances,
		Pools.NodePools(), Disks.InstaceDisksMap())

	p.instances = instances

	instId := set.NewSet()
	instancesMap := map[primitive.ObjectID]*instance.Instance{}
	nodePortsMap := map[string][]*nodeport.Mapping{}
	for _, inst := range instances {
		instId.Add(inst.Id)
		instancesMap[inst.Id] = inst

		nodePortsMap[inst.NetworkNamespace] = append(
			nodePortsMap[inst.NetworkNamespace], inst.NodePorts...)
	}
	p.instancesMap = instancesMap
	p.nodePortsMap = nodePortsMap

	return
}

func (p *InstancesState) Apply(st *State) {
	st.GetInstace = p.GetInstace
	st.Instances = p.Instances
	st.NodePortsMap = p.NodePortsMap
}

func init() {
	InstancesPkg.
		After(InstancesPreload).
		After(Disks).
		After(Pools)
}

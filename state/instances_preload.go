package state

import (
	"sync"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/node"
)

var (
	InstancesPreload    = &InstancesPreloadState{}
	InstancesPreloadPkg = NewPackage(InstancesPreload)
)

type InstancesPreloadState struct {
	roles     []string
	rolesSet  set.Set
	rolesLock sync.Mutex
	instances []*instance.Instance
}

func (p *InstancesPreloadState) GetInstances() []*instance.Instance {
	return p.instances
}

func (p *InstancesPreloadState) GetRoles() (roles []string, rolesSet set.Set) {
	p.rolesLock.Lock()
	roles = p.roles
	rolesSet = p.rolesSet
	p.rolesLock.Unlock()
	return
}

func (p *InstancesPreloadState) setRoles(roles []string, rolesSet set.Set) {
	p.rolesLock.Lock()
	p.roles = roles
	p.rolesSet = rolesSet
	p.rolesLock.Unlock()
	return
}

func (p *InstancesPreloadState) Refresh(pkg *Package,
	db *database.Database) (err error) {

	ndeId := node.Self.Id

	instances, rolesSet, err := instance.GetAllRoles(db, &bson.M{
		"node": ndeId,
	})
	if err != nil {
		return
	}

	p.instances = instances

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

func (p *InstancesPreloadState) Apply(st *State) {
}

package statec

import (
	"sync"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/node"
)

var (
	InstancesPreload    = &InstancesPreloadState{}
	InstancesPreloadPkg = NewPackage(InstancesPreload)
	instancesCache      = NewCache(
		"instances",
		[]string{"node"},
		func() Scope {
			ndeId := node.Self.Id
			return Scope{
				Filter: bson.M{
					"node": ndeId,
				},
				Key: ndeId.Hex(),
				Ok:  true,
			}
		},
		func(evt *Event) bool {
			ndeId, ok := evt.DocObjectId("node")
			return ok && ndeId == node.Self.Id
		},
	).ScopeIds()
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

	err = instancesCache.Sync(db)
	if err != nil {
		return
	}

	instances, err := Decode[instance.Instance](instancesCache)
	if err != nil {
		return
	}

	rolesSet := set.NewSet()
	for _, inst := range instances {
		for _, role := range inst.Roles {
			rolesSet.Add(role)
		}
	}

	p.instances = instances

	nde := node.Self
	if nde.Firewall {
		roles := nde.Roles
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

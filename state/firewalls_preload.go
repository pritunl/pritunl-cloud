package state

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/firewall"
)

var (
	FirewallsPreload    = &FirewallsPreloadState{}
	FirewallsPreloadPkg = NewPackage(FirewallsPreload)
)

type FirewallsPreloadState struct {
	firewalls         map[string][]*firewall.Firewall
	firewallsRolesSet set.Set
}

func (p *FirewallsPreloadState) Firewalls() map[string][]*firewall.Firewall {
	return p.firewalls
}

func (p *FirewallsPreloadState) RolesSet() set.Set {
	return p.firewallsRolesSet
}

func (p *FirewallsPreloadState) Refresh(pkg *Package,
	db *database.Database) (err error) {

	roles, rolesSet := Instances.GetRoles()
	if len(roles) == 0 {
		p.firewalls = map[string][]*firewall.Firewall{}
		p.firewallsRolesSet = set.NewSet()
		return
	}

	firesMap, err := firewall.GetMapRoles(db, roles)
	if err != nil {
		return
	}

	p.firewalls = firesMap
	p.firewallsRolesSet = rolesSet

	return
}

func (p *FirewallsPreloadState) Apply(st *State) {

}

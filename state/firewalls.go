package state

import (
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/firewall"
	"github.com/pritunl/pritunl-cloud/node"
)

var (
	Firewalls    = &FirewallsState{}
	FirewallsPkg = NewPackage(Firewalls)
)

type FirewallsState struct {
	nodeFirewall       []*firewall.Rule
	firewalls          map[string][]*firewall.Rule
	firewallMaps       map[string][]*firewall.Mapping
	instanceNamespaces map[primitive.ObjectID][]string
}

func (p *FirewallsState) NodeFirewall() []*firewall.Rule {
	return p.nodeFirewall
}

func (p *FirewallsState) Firewalls() map[string][]*firewall.Rule {
	return p.firewalls
}

func (p *FirewallsState) FirewallMaps() map[string][]*firewall.Mapping {
	return p.firewallMaps
}

func (p *FirewallsState) GetInstanceNamespaces(
	instId primitive.ObjectID) []string {

	return p.instanceNamespaces[instId]
}

func (p *FirewallsState) Refresh(pkg *Package,
	db *database.Database) (err error) {

	specRules, err := firewall.GetSpecRules(Instances.Instances(),
		Deployments.DeploymentsNode(), Deployments.SpecsMap(),
		Deployments.SpecsUnitsMap(), Deployments.DeploymentsDeployed())
	if err != nil {
		return
	}

	_, rolesSet := InstancesPreload.GetRoles()
	firesMap := FirewallsPreload.Firewalls()
	firewallRolesSet := FirewallsPreload.RolesSet()
	roles := rolesSet.Copy()
	roles.Subtract(firewallRolesSet)

	missRoles := []string{}
	for roleInf := range roles.Iter() {
		missRoles = append(missRoles, roleInf.(string))
	}

	if len(missRoles) > 0 {
		missFiresMap, e := firewall.GetMapRoles(db, missRoles)
		if e != nil {
			err = e
			return
		}

		for role, fires := range missFiresMap {
			firesMap[role] = fires
		}
	}

	nodeFirewall, firewalls, firewallMaps, instNamespaces, err :=
		firewall.GetAllIngressPreloaded(node.Self, Instances.Instances(),
			specRules, Instances.NodePortsMap(), firesMap)
	if err != nil {
		return
	}
	p.nodeFirewall = nodeFirewall
	p.firewalls = firewalls
	p.firewallMaps = firewallMaps
	p.instanceNamespaces = instNamespaces

	return
}

func (p *FirewallsState) Apply(st *State) {
	st.NodeFirewall = p.NodeFirewall
	st.Firewalls = p.Firewalls
	st.FirewallMaps = p.FirewallMaps
	st.GetInstanceNamespaces = p.GetInstanceNamespaces
}

func init() {
	FirewallsPkg.
		After(FirewallsPreload).
		After(Instances).
		After(Vpcs).
		After(Deployments)
}

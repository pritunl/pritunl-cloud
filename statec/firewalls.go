package statec

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
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
	instanceNamespaces map[bson.ObjectID][]string
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
	instId bson.ObjectID) []string {

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
	firesMap := map[string][]*firewall.Firewall{}

	if rolesSet != nil && rolesSet.Len() > 0 {
		fires, e := Decode[firewall.Firewall](firewallsCache)
		if e != nil {
			err = e
			return
		}

		for _, fire := range fires {
			matched := false
			for _, role := range fire.Roles {
				if rolesSet.Contains(role) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}

			for _, role := range fire.Roles {
				roleFires := firesMap[role]
				if roleFires == nil {
					roleFires = []*firewall.Firewall{}
				}
				firesMap[role] = append(roleFires, fire)
			}
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

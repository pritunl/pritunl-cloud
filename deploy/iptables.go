package deploy

import (
	"github.com/pritunl/pritunl-cloud/iptables"
	"github.com/pritunl/pritunl-cloud/state"
)

type Iptables struct {
	stat *state.State
}

func (t *Iptables) Deploy() (err error) {
	nodeSelf := t.stat.Node()
	vpcs := t.stat.Vpcs()
	instaces := t.stat.Instances()
	namespaces := t.stat.Namespaces()
	nodeFirewall := t.stat.NodeFirewall()
	firewalls := t.stat.Firewalls()
	firewallMaps := t.stat.FirewallMaps()

	iptables.UpdateStateRecover(nodeSelf, vpcs, instaces, namespaces,
		nodeFirewall, firewalls, firewallMaps)

	return
}

func NewIptables(stat *state.State) *Iptables {
	return &Iptables{
		stat: stat,
	}
}

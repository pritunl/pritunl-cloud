package deploy

import (
	"github.com/pritunl/pritunl-cloud/ipset"
	"github.com/pritunl/pritunl-cloud/state"
)

type Ipset struct {
	stat *state.State
}

func (t *Ipset) Deploy() (err error) {
	instaces := t.stat.Instances()
	namespaces := t.stat.Namespaces()
	nodeFirewall := t.stat.NodeFirewall()
	firewalls := t.stat.Firewalls()

	err = ipset.UpdateState(instaces, namespaces, nodeFirewall, firewalls)
	if err != nil {
		return
	}

	return
}

func (t *Ipset) Clean() (err error) {
	instaces := t.stat.Instances()
	nodeFirewall := t.stat.NodeFirewall()
	firewalls := t.stat.Firewalls()

	err = ipset.UpdateNamesState(instaces, nodeFirewall, firewalls)
	if err != nil {
		return
	}

	return
}

func NewIpset(stat *state.State) *Ipset {
	return &Ipset{
		stat: stat,
	}
}

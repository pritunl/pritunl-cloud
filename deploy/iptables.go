package deploy

import (
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/iptables"
	"github.com/pritunl/pritunl-cloud/state"
)

type Iptables struct {
	stat *state.State
}

func (t *Iptables) Deploy() (err error) {
	db := database.GetDatabase()
	defer db.Close()

	nodeSelf := t.stat.Node()
	instaces := t.stat.Instances()
	namespaces := t.stat.Namespaces()
	nodeFirewall := t.stat.NodeFirewall()
	firewalls := t.stat.Firewalls()

	iptables.UpdateStateRecover(nodeSelf, instaces, namespaces,
		nodeFirewall, firewalls)

	return
}

func NewIptables(stat *state.State) *Iptables {
	return &Iptables{
		stat: stat,
	}
}

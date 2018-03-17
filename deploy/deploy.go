package deploy

import (
	"github.com/pritunl/pritunl-cloud/state"
)

func Deploy(stat *state.State) (err error) {
	iptables := NewIptables(stat)
	err = iptables.Deploy()
	if err != nil {
		return
	}

	disks := NewDisks(stat)
	err = disks.Deploy()
	if err != nil {
		return
	}

	instances := NewInstances(stat)
	err = instances.Deploy()
	if err != nil {
		return
	}

	namespaces := NewNamespace(stat)
	err = namespaces.Deploy()
	if err != nil {
		return
	}

	return
}

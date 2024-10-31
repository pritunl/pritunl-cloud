package deploy

import (
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/state"
)

func Deploy(stat *state.State) (err error) {
	db := database.GetDatabase()
	defer db.Close()

	network := NewNetwork(stat)
	err = network.Deploy()
	if err != nil {
		return
	}

	ipset := NewIpset(stat)
	err = ipset.Deploy()
	if err != nil {
		return
	}

	iptables := NewIptables(stat)
	err = iptables.Deploy()
	if err != nil {
		return
	}

	err = ipset.Clean()
	if err != nil {
		return
	}

	disks := NewDisks(stat)
	err = disks.Deploy(db)
	if err != nil {
		return
	}

	instances := NewInstances(stat)
	err = instances.Deploy(db)
	if err != nil {
		return
	}

	namespaces := NewNamespace(stat)
	err = namespaces.Deploy()
	if err != nil {
		return
	}

	services := NewServices(stat)
	err = services.Deploy(db)
	if err != nil {
		return
	}

	deployments := NewDeployments(stat)
	err = deployments.Deploy(db)
	if err != nil {
		return
	}

	imds := NewImds(stat)
	err = imds.Deploy(db)
	if err != nil {
		return
	}

	return
}

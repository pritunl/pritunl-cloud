package deploy

import (
	"time"

	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/state"
)

func Deploy(stat *state.State, runtimes *state.Runtimes) (err error) {
	db := database.GetDatabase()
	defer db.Close()

	start := time.Now()
	network := NewNetwork(stat)
	err = network.Deploy()
	if err != nil {
		return
	}
	runtimes.Network = time.Since(start)

	start = time.Now()
	ipset := NewIpset(stat)
	err = ipset.Deploy()
	if err != nil {
		return
	}
	runtimes.Ipset = time.Since(start)

	start = time.Now()
	iptables := NewIptables(stat)
	err = iptables.Deploy()
	if err != nil {
		return
	}

	err = ipset.Clean()
	if err != nil {
		return
	}
	runtimes.Iptables = time.Since(start)

	start = time.Now()
	disks := NewDisks(stat)
	err = disks.Deploy(db)
	if err != nil {
		return
	}
	runtimes.Disks = time.Since(start)

	start = time.Now()
	instances := NewInstances(stat)
	err = instances.Deploy(db)
	if err != nil {
		return
	}
	runtimes.Instances = time.Since(start)

	start = time.Now()
	namespaces := NewNamespace(stat)
	err = namespaces.Deploy(db)
	if err != nil {
		return
	}
	runtimes.Namespaces = time.Since(start)

	start = time.Now()
	pods := NewPods(stat)
	err = pods.Deploy(db)
	if err != nil {
		return
	}
	runtimes.Pods = time.Since(start)

	start = time.Now()
	deployments := NewDeployments(stat)
	err = deployments.Deploy(db)
	if err != nil {
		return
	}
	runtimes.Deployments = time.Since(start)

	start = time.Now()
	imds := NewImds(stat)
	err = imds.Deploy(db)
	if err != nil {
		return
	}
	runtimes.Imds = time.Since(start)

	start = time.Now()
	stat.Wait()
	runtimes.Wait = time.Since(start)

	return
}

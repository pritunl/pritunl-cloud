package deploy

import (
	"strings"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/interfaces"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/state"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
)

var (
	firstRun         = true
	namespaceLock    = utils.NewMultiTimeoutLock(5 * time.Minute)
	namespaceLimiter = utils.NewLimiter(5)
)

type Namespace struct {
	stat *state.State
}

func (n *Namespace) Deploy(db *database.Database) (err error) {
	instances := n.stat.Instances()
	namespaces := n.stat.Namespaces()
	ifaces := n.stat.Interfaces()

	curNamespaces := set.NewSet()
	curVirtIfaces := set.NewSet()
	curExternalIfaces := set.NewSet()

	nodeNetworkMode := node.Self.NetworkMode
	if nodeNetworkMode == "" {
		nodeNetworkMode = node.Dhcp
	}

	nodeNetworkMode6 := node.Self.NetworkMode6
	if nodeNetworkMode6 == "" {
		nodeNetworkMode6 = node.Dhcp
	}

	externalNetwork := false
	if (nodeNetworkMode != node.Disabled &&
		nodeNetworkMode != node.Oracle) ||
		(nodeNetworkMode6 != node.Disabled &&
			nodeNetworkMode6 != node.Oracle) {

		externalNetwork = true
	}

	for _, inst := range instances {
		if !inst.IsActive() {
			continue
		}

		curNamespaces.Add(vm.GetNamespace(inst.Id, 0))
		if externalNetwork {
			curVirtIfaces.Add(vm.GetIfaceNodeExternal(inst.Id, 0))
		}
		curVirtIfaces.Add(vm.GetIfaceNodeInternal(inst.Id, 0))
		curVirtIfaces.Add(vm.GetIfaceHost(inst.Id, 0))
		if externalNetwork {
			curExternalIfaces.Add(vm.GetIfaceExternal(inst.Id, 0))
		}
		curVirtIfaces.Add(vm.GetIfaceNodePort(inst.Id, 0))
	}

	firstRun = false

	for _, iface := range ifaces {
		if len(iface) != 14 || !(strings.HasPrefix(iface, "j") ||
			strings.HasPrefix(iface, "r") ||
			utils.HasPreSuf(iface, "h", "0") ||
			utils.HasPreSuf(iface, "m", "0")) {

			continue
		}

		if !curVirtIfaces.Contains(iface) {
			utils.ExecCombinedOutputLogged(
				[]string{
					"Cannot find device",
				},
				"ip", "link", "del", iface,
			)
			interfaces.RemoveVirtIface(iface)
		}
	}

	for _, namespace := range namespaces {
		if len(namespace) != 14 || !strings.HasPrefix(namespace, "n") {
			continue
		}

		if !curNamespaces.Contains(namespace) {
			_, err = utils.ExecCombinedOutputLogged(
				[]string{
					"No such file",
				},
				"ip", "netns", "del", namespace,
			)
			if err != nil {
				return
			}
		}
	}

	return
}

func NewNamespace(stat *state.State) *Namespace {
	return &Namespace{
		stat: stat,
	}
}

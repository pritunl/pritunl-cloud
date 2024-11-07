package deploy

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/interfaces"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/state"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
)

var (
	firstRun = true
)

type Namespace struct {
	stat *state.State
}

func (n *Namespace) Deploy() (err error) {
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

	hostNetwork := false
	if !node.Self.HostBlock.IsZero() {
		hostNetwork = true
	}

	for _, inst := range instances {
		if !inst.IsActive() {
			continue
		}

		curNamespaces.Add(vm.GetNamespace(inst.Id, 0))
		if externalNetwork {
			curVirtIfaces.Add(vm.GetIfaceVirt(inst.Id, 0))
		}
		curVirtIfaces.Add(vm.GetIfaceVirt(inst.Id, 1))
		if hostNetwork {
			curVirtIfaces.Add(vm.GetIfaceVirt(inst.Id, 2))
		}
		if externalNetwork {
			curExternalIfaces.Add(vm.GetIfaceExternal(inst.Id, 0))
		}

		// TODO Upgrade code
		if firstRun {
			namespace := vm.GetNamespace(inst.Id, 0)
			iface := vm.GetIfaceExternal(inst.Id, 1)

			_, _ = utils.ExecCombinedOutput("",
				"ip", "netns", "exec", namespace,
				"ip", "link", "set", iface, "down")
			_, _ = utils.ExecCombinedOutput("",
				"ip", "netns", "exec", namespace,
				"ip", "link", "del", iface)
		}
	}

	firstRun = false

	for _, iface := range ifaces {
		if len(iface) != 14 || !strings.HasPrefix(iface, "v") {
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

	running := n.stat.Running()
	for _, name := range running {
		if len(name) != 27 || !strings.HasPrefix(name, "dhclient-i") {
			continue
		}

		iface := name[9:23]

		if !curExternalIfaces.Contains(iface) {
			pth := filepath.Join("/var/run", name)

			pidByt, e := ioutil.ReadFile(pth)
			if e != nil {
				err = &errortypes.ReadError{
					errors.Wrap(e, "namespace: Failed to read dhclient pid"),
				}
				return
			}

			pid, e := strconv.Atoi(strings.TrimSpace(string(pidByt)))
			if e != nil {
				err = &errortypes.ParseError{
					errors.Wrap(e, "namespace: Failed to parse dhclient pid"),
				}
				return
			}

			exists, _ := utils.Exists(fmt.Sprintf("/proc/%d/status", pid))
			if exists {
				utils.ExecCombinedOutput("", "kill", "-9", strconv.Itoa(pid))
			} else {
				os.Remove(pth)
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

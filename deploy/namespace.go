package deploy

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/interfaces"
	"github.com/pritunl/pritunl-cloud/netconf"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/state"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/sirupsen/logrus"
)

var (
	firstRun         = true
	namespaceLock    = utils.NewMultiTimeoutLock(5 * time.Minute)
	namespaceLimiter = utils.NewLimiter(5)
)

type Namespace struct {
	stat *state.State
}

func (n *Namespace) restartDhcp(inst *instance.Instance) {
	virt := n.stat.GetVirt(inst.Id)
	if virt == nil {
		return
	}
	nc := netconf.New(virt)

	if !namespaceLimiter.Acquire() {
		return
	}

	acquired, lockId := namespaceLock.LockOpenTimeout(
		inst.Id.Hex(), 10*time.Minute)
	if !acquired {
		namespaceLimiter.Release()
		return
	}

	go func() {
		defer func() {
			time.Sleep(3 * time.Second)
			namespaceLock.Unlock(inst.Id.Hex(), lockId)
			namespaceLimiter.Release()
		}()

		db := database.GetDatabase()
		defer db.Close()

		logrus.WithFields(logrus.Fields{
			"instance_id": inst.Id.Hex(),
		}).Debug("deploy: Restarting instance dhclient6")

		err := nc.RestartDhcp6(db)
		if err != nil {
			return
		}
	}()
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

	now := time.Now()
	nde := n.stat.Node()
	hasDhcp6 := nde.NetworkMode6 == node.Dhcp ||
		nde.NetworkMode6 == node.DhcpSlaac
	dhcpTtl := settings.Hypervisor.Dhcp6RefreshTtl
	dhcpRefresh := time.Duration(dhcpTtl) * time.Second

	if hasDhcp6 && !netconf.DhTimestampsLoaded {
		err = netconf.LoadDhTimestamps(instances)
		if err != nil {
			return
		}
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

		if hasDhcp6 && dhcpTtl != 0 {
			timestamp := netconf.GetDhTimestamp(inst.Id)
			if now.Sub(timestamp) > dhcpRefresh {
				n.restartDhcp(inst)
			}
		}
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

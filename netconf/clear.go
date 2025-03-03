package netconf

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/interfaces"
	"github.com/pritunl/pritunl-cloud/store"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/pritunl/tools/commander"
)

func (n *NetConf) Clear(db *database.Database) (err error) {
	clearIface("", n.SystemExternalIface)
	clearIface("", n.SystemInternalIface)
	clearIface("", n.SystemHostIface)
	clearIface("", n.SystemNodePortIface)
	clearIface("", n.SpaceExternalIface)
	clearIface("", n.SpaceInternalIface)
	clearIface("", n.SpaceHostIface)
	clearIface("", n.SpaceNodePortIface)

	clearIface(n.Namespace, n.SpaceBridgeIface)
	clearIface(n.Namespace, n.SpaceImdsIface)

	interfaces.RemoveVirtIface(n.SystemExternalIface)
	interfaces.RemoveVirtIface(n.SystemInternalIface)
	interfaces.RemoveVirtIface(n.SystemNodePortIface)

	return
}

func (n *NetConf) ClearAll(db *database.Database) (err error) {
	if len(n.Virt.NetworkAdapters) == 0 {
		err = &errortypes.NotFoundError{
			errors.New("qemu: Missing network interfaces"),
		}
		return
	}

	ifaceExternal := vm.GetIfaceExternal(n.Virt.Id, 0)
	pidPath := fmt.Sprintf(n.DhcpPidPath, ifaceExternal)

	pid := ""
	pidData, _ := ioutil.ReadFile(pidPath)
	if pidData != nil {
		pid = strings.TrimSpace(string(pidData))
	}

	if pid != "" {
		_, _ = utils.ExecCombinedOutput("", "kill", pid)
	}

	pidPath = fmt.Sprintf(n.Dhcp6PidPath, ifaceExternal)

	pid = ""
	pidData, _ = ioutil.ReadFile(pidPath)
	if pidData != nil {
		pid = strings.TrimSpace(string(pidData))
	}

	if pid != "" {
		_, _ = utils.ExecCombinedOutput("", "kill", pid)
	}

	_ = utils.RemoveAll(pidPath)

	err = n.Clear(db)
	if err != nil {
		return
	}

	store.RemAddress(n.Virt.Id)
	store.RemRoutes(n.Virt.Id)
	store.RemArp(n.Virt.Id)

	return
}

func clearIface(namespace, iface string) {
	if iface == "" {
		return
	}

	if namespace != "" {
		commander.Exec(&commander.Opt{
			Name: "ip",
			Args: []string{
				"netns", "exec", namespace,
				"ip", "link", "set", iface, "down",
			},
			PipeOut: true,
			PipeErr: true,
		})
		commander.Exec(&commander.Opt{
			Name: "ip",
			Args: []string{
				"netns", "exec", namespace,
				"ip", "link", "del", iface,
			},
			PipeOut: true,
			PipeErr: true,
		})
	} else {
		commander.Exec(&commander.Opt{
			Name: "ip",
			Args: []string{
				"link", "set", iface, "down",
			},
			PipeOut: true,
			PipeErr: true,
		})
		commander.Exec(&commander.Opt{
			Name: "ip",
			Args: []string{
				"link", "del", iface,
			},
			PipeOut: true,
			PipeErr: true,
		})
	}

}

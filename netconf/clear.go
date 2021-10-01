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
)

func (n *NetConf) Clear(db *database.Database) (err error) {
	_, err = utils.ExecCombinedOutputLogged(
		[]string{"File exists"},
		"ip", "netns",
		"add", n.Namespace,
	)
	if err != nil {
		return
	}

	clearIface(n.SystemExternalIface)
	clearIface(n.SystemExternalIface6)
	clearIface(n.SystemInternalIface)
	clearIface(n.SystemHostIface)
	clearIface(n.SpaceExternalIface)
	clearIface(n.SpaceExternalIface6)
	clearIface(n.SpaceInternalIface)
	clearIface(n.SpaceHostIface)

	interfaces.RemoveVirtIface(n.SystemExternalIface)
	interfaces.RemoveVirtIface(n.SystemExternalIface6)
	interfaces.RemoveVirtIface(n.SystemInternalIface)

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
	pidPath := fmt.Sprintf("/var/run/dhclient-%s.pid", ifaceExternal)

	pid := ""
	pidData, _ := ioutil.ReadFile(pidPath)
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

	return
}

func clearIface(iface string) {
	_, _ = utils.ExecCombinedOutput(
		"", "ip", "link", "set", iface, "down")
	_, _ = utils.ExecCombinedOutput(
		"", "ip", "link", "del", iface)
}

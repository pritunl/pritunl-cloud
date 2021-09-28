package netconf

import (
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/utils"
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

	// TODO Clear br0 namespace interface
	clearIface(n.SystemExternalIface)
	clearIface(n.SystemExternalIface6)
	clearIface(n.SystemInternalIface)
	clearIface(n.SystemHostIface)
	clearIface(n.SpaceExternalIface)
	clearIface(n.SpaceExternalIface6)
	clearIface(n.SpaceInternalIface)
	clearIface(n.SpaceHostIface)

	return
}

func clearIface(iface string) {
	_, _ = utils.ExecCombinedOutput(
		"", "ip", "link", "set", iface, "down")
	_, _ = utils.ExecCombinedOutput(
		"", "ip", "link", "del", iface)
}

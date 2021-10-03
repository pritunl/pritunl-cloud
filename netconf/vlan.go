package netconf

import (
	"strconv"

	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/utils"
)

func (n *NetConf) vlanNet(db *database.Database) (err error) {
	_, err = utils.ExecCombinedOutputLogged(
		[]string{"File exists"},
		"ip", "netns", "exec", n.Namespace,
		"ip", "link",
		"add", "link", n.SpaceInternalIface,
		"name", n.BridgeInternalIface,
		"type", "vlan",
		"id", strconv.Itoa(n.VlanId),
	)
	if err != nil {
		return
	}

	return
}

func (n *NetConf) vlanMtu(db *database.Database) (err error) {
	if n.SpaceInternalIfaceMtu != "" {
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "netns", "exec", n.Namespace,
			"ip", "link",
			"set", "dev", n.BridgeInternalIface,
			"mtu", n.SpaceInternalIfaceMtu,
		)
		if err != nil {
			return
		}
	}

	return
}

func (n *NetConf) vlanUp(db *database.Database) (err error) {
	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", n.Namespace,
		"ip", "link",
		"set", "dev", n.BridgeInternalIface, "up",
	)
	if err != nil {
		return
	}

	return
}

func (n *NetConf) Vlan(db *database.Database) (err error) {
	err = n.vlanNet(db)
	if err != nil {
		return
	}

	err = n.vlanMtu(db)
	if err != nil {
		return
	}

	err = n.vlanUp(db)
	if err != nil {
		return
	}

	return
}

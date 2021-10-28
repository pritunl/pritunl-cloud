package netconf

import (
	"fmt"

	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/utils"
)

func (n *NetConf) externalNet(db *database.Database) (err error) {
	if n.NetworkMode != node.Disabled && n.NetworkMode != node.Oracle {
		if n.PhysicalExternalIfaceBridge {
			_, err = utils.ExecCombinedOutputLogged(
				nil,
				"ip", "link",
				"add", n.SystemExternalIface,
				"type", "veth",
				"peer", "name", n.SpaceExternalIface,
				"addr", n.ExternalMacAddr,
			)
			if err != nil {
				return
			}
		} else {
			_, err = utils.ExecCombinedOutputLogged(
				nil,
				"ip", "link",
				"add", n.SpaceExternalIface,
				"addr", n.ExternalMacAddr,
				"link", n.PhysicalExternalIface,
				"type", "macvlan",
				"mode", "bridge",
			)
			if err != nil {
				return
			}
		}
	}

	if n.NetworkMode6 != node.Disabled &&
		n.SpaceExternalIface != n.SpaceExternalIface6 &&
		n.NetworkMode6 != node.Oracle {

		if n.PhysicalExternalIfaceBridge6 {
			_, err = utils.ExecCombinedOutputLogged(
				nil,
				"ip", "link",
				"add", n.SystemExternalIface6,
				"type", "veth",
				"peer", "name", n.SpaceExternalIface6,
				"addr", n.ExternalMacAddr6,
			)
			if err != nil {
				return
			}
		} else {
			_, err = utils.ExecCombinedOutputLogged(
				nil,
				"ip", "link",
				"add", n.SpaceExternalIface6,
				"addr", n.ExternalMacAddr6,
				"link", n.PhysicalExternalIface6,
				"type", "macvlan",
				"mode", "bridge",
			)
			if err != nil {
				return
			}
		}
	}

	return
}

func (n *NetConf) externalMtu(db *database.Database) (err error) {
	if n.NetworkMode != node.Disabled &&
		n.PhysicalExternalIfaceBridge &&
		n.SystemExternalIfaceMtu != "" &&
		n.NetworkMode != node.Oracle {

		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "link",
			"set", "dev", n.SystemExternalIface,
			"mtu", n.SystemExternalIfaceMtu,
		)
		if err != nil {
			return
		}
	}
	if n.NetworkMode6 != node.Disabled &&
		n.PhysicalExternalIfaceBridge6 &&
		n.SystemExternalIfaceMtu6 != "" &&
		n.NetworkMode6 != node.Oracle {

		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "link",
			"set", "dev", n.SystemExternalIface6,
			"mtu", n.SystemExternalIfaceMtu6,
		)
		if err != nil {
			return
		}
	}

	if n.NetworkMode != node.Disabled && n.SpaceExternalIfaceMtu != "" &&
		n.NetworkMode != node.Oracle {

		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "link",
			"set", "dev", n.SpaceExternalIface,
			"mtu", n.SpaceExternalIfaceMtu,
		)
		if err != nil {
			return
		}
	}
	if n.NetworkMode6 != node.Disabled && n.SpaceExternalIfaceMtu6 != "" &&
		n.NetworkMode != node.Oracle {

		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "link",
			"set", "dev", n.SpaceExternalIface6,
			"mtu", n.SpaceExternalIfaceMtu6,
		)
		if err != nil {
			return
		}
	}

	return
}

func (n *NetConf) externalUp(db *database.Database) (err error) {
	if n.NetworkMode != node.Disabled && n.PhysicalExternalIfaceBridge &&
		n.NetworkMode != node.Oracle {

		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "link",
			"set", "dev", n.SystemExternalIface, "up",
		)
		if err != nil {
			return
		}
	}

	if n.NetworkMode6 != node.Disabled && n.PhysicalExternalIfaceBridge6 &&
		n.SystemExternalIface != n.SystemExternalIface6 &&
		n.NetworkMode6 != node.Oracle {

		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "link",
			"set", "dev", n.SystemExternalIface6, "up",
		)
		if err != nil {
			return
		}
	}

	return
}

func (n *NetConf) externalSysctl(db *database.Database) (err error) {
	if n.NetworkMode6 != node.Disabled && n.NetworkMode6 != node.Oracle {
		_, err = utils.ExecCombinedOutputLogged(
			nil, "sysctl", "-w",
			fmt.Sprintf("net.ipv6.conf.%s.accept_ra=2",
				n.PhysicalExternalIface6),
		)
		if err != nil {
			return
		}
	}

	return
}

func (n *NetConf) externalMaster(db *database.Database) (err error) {
	if n.NetworkMode != node.Disabled && n.PhysicalExternalIfaceBridge &&
		n.NetworkMode != node.Oracle {

		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "link",
			"set", n.SystemExternalIface,
			"master", n.PhysicalExternalIface,
		)
		if err != nil {
			return
		}
	}

	if n.NetworkMode6 != node.Disabled && n.PhysicalExternalIfaceBridge6 &&
		n.SystemExternalIface != n.SystemExternalIface6 &&
		n.NetworkMode6 != node.Oracle {

		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "link",
			"set", n.SystemExternalIface6,
			"master", n.PhysicalExternalIface6,
		)
		if err != nil {
			return
		}
	}

	return
}

func (n *NetConf) externalSpace(db *database.Database) (err error) {
	if n.NetworkMode != node.Disabled && n.NetworkMode != node.Oracle {
		_, err = utils.ExecCombinedOutputLogged(
			[]string{"File exists"},
			"ip", "link",
			"set", "dev", n.SpaceExternalIface,
			"netns", n.Namespace,
		)
		if err != nil {
			return
		}
	}

	if n.NetworkMode6 != node.Disabled &&
		n.SpaceExternalIface != n.SpaceExternalIface6 &&
		n.NetworkMode6 != node.Oracle {

		_, err = utils.ExecCombinedOutputLogged(
			[]string{"File exists"},
			"ip", "link",
			"set", "dev", n.SpaceExternalIface6,
			"netns", n.Namespace,
		)
		if err != nil {
			return
		}
	}

	return
}

func (n *NetConf) externalSpaceUp(db *database.Database) (err error) {
	if n.NetworkMode != node.Disabled && n.NetworkMode != node.Oracle {
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "netns", "exec", n.Namespace,
			"ip", "link",
			"set", "dev", n.SpaceExternalIface, "up",
		)
		if err != nil {
			return
		}
	}

	if n.NetworkMode6 != node.Disabled &&
		n.SpaceExternalIface != n.SpaceExternalIface6 &&
		n.NetworkMode6 != node.Oracle {

		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "netns", "exec", n.Namespace,
			"ip", "link",
			"set", "dev", n.SpaceExternalIface6, "up",
		)
		if err != nil {
			return
		}
	}

	return
}

func (n *NetConf) External(db *database.Database) (err error) {
	err = n.externalNet(db)
	if err != nil {
		return
	}

	err = n.externalMtu(db)
	if err != nil {
		return
	}

	err = n.externalUp(db)
	if err != nil {
		return
	}

	err = n.externalSysctl(db)
	if err != nil {
		return
	}

	err = n.externalMaster(db)
	if err != nil {
		return
	}

	err = n.externalSpace(db)
	if err != nil {
		return
	}

	err = n.externalSpaceUp(db)
	if err != nil {
		return
	}

	return
}

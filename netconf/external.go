package netconf

import (
	"fmt"
	"strconv"

	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/utils"
)

func (n *NetConf) externalNet(db *database.Database) (err error) {
	if (n.NetworkMode != node.Disabled && n.NetworkMode != node.Oracle) ||
		(n.NetworkMode6 != node.Disabled && n.NetworkMode6 != node.Oracle) {

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

	return
}

func (n *NetConf) externalMtu(db *database.Database) (err error) {
	if (n.PhysicalExternalIfaceBridge &&
		n.SystemExternalIfaceMtu != "") &&
		((n.NetworkMode != node.Disabled &&
			n.NetworkMode != node.Oracle) ||
			(n.NetworkMode6 != node.Disabled &&
				n.NetworkMode6 != node.Oracle)) {

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

	if n.SpaceExternalIfaceMtu != "" &&
		((n.NetworkMode != node.Disabled &&
			n.NetworkMode != node.Oracle) ||
			(n.NetworkMode6 != node.Disabled &&
				n.NetworkMode6 != node.Oracle)) {

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

	return
}

func (n *NetConf) externalUp(db *database.Database) (err error) {
	if n.PhysicalExternalIfaceBridge &&
		((n.NetworkMode != node.Disabled &&
			n.NetworkMode != node.Oracle) ||
			(n.NetworkMode6 != node.Disabled &&
				n.NetworkMode6 != node.Oracle)) {

		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "link",
			"set", "dev", n.SystemExternalIface, "up",
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
				n.PhysicalExternalIface),
		)
		if err != nil {
			return
		}

		if n.NetworkMode6 != node.Slaac && n.NetworkMode6 != node.DhcpSlaac {
			_, err = utils.ExecCombinedOutputLogged(
				nil, "sysctl", "-w",
				fmt.Sprintf("net.ipv6.conf.%s.addr_gen_mode=1",
					n.PhysicalExternalIface),
			)
			if err != nil {
				return
			}
		}
	}

	return
}

func (n *NetConf) externalMaster(db *database.Database) (err error) {
	if n.PhysicalExternalIfaceBridge &&
		((n.NetworkMode != node.Disabled &&
			n.NetworkMode != node.Oracle) ||
			(n.NetworkMode6 != node.Disabled &&
				n.NetworkMode6 != node.Oracle)) {

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

	return
}

func (n *NetConf) externalSpace(db *database.Database) (err error) {
	if (n.NetworkMode != node.Disabled && n.NetworkMode != node.Oracle) ||
		(n.NetworkMode6 != node.Disabled && n.NetworkMode6 != node.Oracle) {

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

	return
}

func (n *NetConf) externalSpaceMod(db *database.Database) (err error) {
	if n.SpaceExternalIfaceMod != "" {
		_, err = utils.ExecCombinedOutputLogged(
			[]string{"File exists"},
			"ip", "netns", "exec", n.Namespace,
			"ip", "link",
			"add", "link", n.SpaceExternalIface,
			"name", n.SpaceExternalIfaceMod,
			"type", "vlan",
			"id", strconv.Itoa(n.ExternalVlan),
		)
		if err != nil {
			return
		}

		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "netns", "exec", n.Namespace,
			"ip", "link",
			"set", "dev", n.SpaceExternalIfaceMod,
			"mtu", n.SpaceExternalIfaceMtu,
		)
		if err != nil {
			return
		}
	}

	if n.SpaceExternalIfaceMod6 != "" &&
		n.SpaceExternalIfaceMod6 != n.SpaceExternalIfaceMod {

		_, err = utils.ExecCombinedOutputLogged(
			[]string{"File exists"},
			"ip", "netns", "exec", n.Namespace,
			"ip", "link",
			"add", "link", n.SpaceExternalIface,
			"name", n.SpaceExternalIfaceMod6,
			"type", "vlan",
			"id", strconv.Itoa(n.ExternalVlan6),
		)
		if err != nil {
			return
		}

		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "netns", "exec", n.Namespace,
			"ip", "link",
			"set", "dev", n.SpaceExternalIfaceMod6,
			"mtu", n.SpaceExternalIfaceMtu,
		)
		if err != nil {
			return
		}
	}

	return
}

func (n *NetConf) externalSpaceSysctl(db *database.Database) (err error) {
	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", n.Namespace,
		"sysctl", "-w", "net.ipv6.conf.all.accept_ra=0",
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", n.Namespace,
		"sysctl", "-w", "net.ipv6.conf.default.accept_ra=0",
	)
	if err != nil {
		return
	}

	if n.NetworkMode6 == node.Static {
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "netns", "exec", n.Namespace,
			"sysctl", "-w",
			fmt.Sprintf("net.ipv6.conf.%s.autoconf=0",
				n.SpaceExternalIface),
		)
		if err != nil {
			return
		}

		if n.SpaceExternalIfaceMod6 != "" {
			_, err = utils.ExecCombinedOutputLogged(
				nil,
				"ip", "netns", "exec", n.Namespace,
				"sysctl", "-w",
				fmt.Sprintf("net.ipv6.conf.%s.autoconf=0",
					n.SpaceExternalIfaceMod6),
			)
			if err != nil {
				return
			}
		}
	}

	if n.NetworkMode6 != node.Disabled && n.NetworkMode6 != node.Oracle {
		if n.SpaceExternalIfaceMod6 == "" {
			_, err = utils.ExecCombinedOutputLogged(
				nil,
				"ip", "netns", "exec", n.Namespace,
				"sysctl", "-w",
				fmt.Sprintf("net.ipv6.conf.%s.accept_ra=2",
					n.SpaceExternalIface),
			)
			if err != nil {
				return
			}
		} else {
			_, err = utils.ExecCombinedOutputLogged(
				nil,
				"ip", "netns", "exec", n.Namespace,
				"sysctl", "-w",
				fmt.Sprintf("net.ipv6.conf.%s.accept_ra=2",
					n.SpaceExternalIfaceMod6),
			)
			if err != nil {
				return
			}
		}
	}

	if (n.NetworkMode != node.Disabled && n.NetworkMode != node.Oracle) &&
		(n.NetworkMode6 == node.Disabled || n.NetworkMode6 == node.Oracle) {

		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "netns", "exec", n.Namespace,
			"sysctl", "-w",
			fmt.Sprintf("net.ipv6.conf.%s.disable_ipv6=1",
				n.SpaceExternalIface),
		)
		if err != nil {
			return
		}
	}

	if n.SpaceExternalIfaceMod != n.SpaceExternalIfaceMod6 &&
		n.SpaceExternalIfaceMod != "" {

		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "netns", "exec", n.Namespace,
			"sysctl", "-w",
			fmt.Sprintf("net.ipv6.conf.%s.disable_ipv6=1",
				n.SpaceExternalIfaceMod),
		)
		if err != nil {
			return
		}
	}

	return
}

func (n *NetConf) externalSpaceUp(db *database.Database) (err error) {
	if (n.NetworkMode != node.Disabled && n.NetworkMode != node.Oracle) ||
		(n.NetworkMode6 != node.Disabled && n.NetworkMode6 != node.Oracle) {

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

	if n.SpaceExternalIfaceMod != "" {
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "netns", "exec", n.Namespace,
			"ip", "link",
			"set", "dev", n.SpaceExternalIfaceMod, "up",
		)
		if err != nil {
			return
		}
	}

	if n.SpaceExternalIfaceMod6 != "" &&
		n.SpaceExternalIfaceMod6 != n.SpaceExternalIfaceMod {

		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "netns", "exec", n.Namespace,
			"ip", "link",
			"set", "dev", n.SpaceExternalIfaceMod6, "up",
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

	err = n.externalSpaceMod(db)
	if err != nil {
		return
	}

	err = n.externalSpaceSysctl(db)
	if err != nil {
		return
	}

	err = n.externalSpaceUp(db)
	if err != nil {
		return
	}

	return
}

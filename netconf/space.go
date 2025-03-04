package netconf

import (
	"fmt"

	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/iptables"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/utils"
)

func (n *NetConf) spaceSysctl(db *database.Database) (err error) {
	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", n.Namespace,
		"sysctl", "-w", "net.ipv4.conf.all.accept_redirects=0",
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", n.Namespace,
		"sysctl", "-w", "net.ipv4.conf.default.accept_redirects=0",
	)
	if err != nil {
		return
	}
	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", n.Namespace,
		"sysctl", "-w", "net.ipv4.conf.all.rp_filter=1",
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", n.Namespace,
		"sysctl", "-w", "net.ipv4.conf.default.rp_filter=1",
	)
	if err != nil {
		return
	}

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
	}

	if n.NetworkMode6 != node.Disabled && n.NetworkMode6 != node.Oracle {
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
	}

	if n.HostNetwork {
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "netns", "exec", n.Namespace,
			"sysctl", "-w",
			fmt.Sprintf("net.ipv6.conf.%s.disable_ipv6=1",
				n.SpaceHostIface),
		)
		if err != nil {
			return
		}
	}

	if n.NodePortNetwork {
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "netns", "exec", n.Namespace,
			"sysctl", "-w",
			fmt.Sprintf("net.ipv6.conf.%s.disable_ipv6=1",
				n.SpaceNodePortIface),
		)
		if err != nil {
			return
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

	return
}

func (n *NetConf) spaceForward(db *database.Database) (err error) {
	if (n.NetworkMode != node.Disabled && n.NetworkMode != node.Oracle) ||
		(n.NetworkMode6 != node.Disabled && n.NetworkMode6 != node.Oracle) {

		_, err = utils.ExecCombinedOutputLogged(
			[]string{"already exists"},
			"ip", "netns", "exec", n.Namespace,
			"ipset",
			"create", "prx_inst6", "hash:net",
			"family", "inet6",
		)
		if err != nil {
			return
		}

		_, err = utils.ExecCombinedOutputLogged(
			[]string{"already added"},
			"ip", "netns", "exec", n.Namespace,
			"ipset",
			"add", "prx_inst6", "fe80::/64",
		)
		if err != nil {
			return
		}

		_, err = utils.ExecCombinedOutputLogged(
			[]string{"already added"},
			"ip", "netns", "exec", n.Namespace,
			"ipset",
			"add", "prx_inst6", n.InternalAddr6.String()+"/128",
		)
		if err != nil {
			return
		}

		iptables.Lock()
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "netns", "exec", n.Namespace,
			"iptables",
			"-I", "FORWARD", "1",
			"!", "-d", n.InternalAddr.String()+"/32",
			"-i", n.SpaceExternalIface,
			"-m", "comment",
			"--comment", "pritunl_cloud_base",
			"-j", "DROP",
		)
		iptables.Unlock()
		if err != nil {
			return
		}

		iptables.Lock()
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "netns", "exec", n.Namespace,
			"ip6tables",
			"-I", "FORWARD", "1",
			"-m", "set",
			"!", "--match-set", "prx_inst6", "dst",
			"-i", n.SpaceExternalIface,
			"-m", "comment",
			"--comment", "pritunl_cloud_base",
			"-j", "DROP",
		)
		iptables.Unlock()
		if err != nil {
			return
		}
	}

	if n.HostNetwork {
		iptables.Lock()
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "netns", "exec", n.Namespace,
			"iptables",
			"-I", "FORWARD", "1",
			"!", "-d", n.InternalAddr.String()+"/32",
			"-i", n.SpaceHostIface,
			"-m", "comment",
			"--comment", "pritunl_cloud_base",
			"-j", "DROP",
		)
		iptables.Unlock()
		if err != nil {
			return
		}
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", n.Namespace,
		"sysctl", "-w", "net.ipv4.ip_forward=1",
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", n.Namespace,
		"sysctl", "-w", "net.ipv6.conf.all.forwarding=1",
	)
	if err != nil {
		return
	}

	return
}

func (n *NetConf) spaceVirt(db *database.Database) (err error) {
	_, err = utils.ExecCombinedOutputLogged(
		[]string{
			"Cannot find device",
			"File exists",
		},
		"ip", "link",
		"set", "dev", n.VirtIface,
		"netns", n.Namespace,
	)
	if err != nil {
		return
	}

	return
}

func (n *NetConf) spaceLoopback(db *database.Database) (err error) {
	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", n.Namespace,
		"ip", "link",
		"set", "dev", "lo", "up",
	)
	if err != nil {
		return
	}

	return
}

func (n *NetConf) spaceMtu(db *database.Database) (err error) {
	if n.VirtIfaceMtu != "" {
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "netns", "exec", n.Namespace,
			"ip", "link",
			"set", "dev", n.VirtIface,
			"mtu", n.VirtIfaceMtu,
		)
		if err != nil {
			return
		}
	}

	return
}

func (n *NetConf) spaceUp(db *database.Database) (err error) {
	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", n.Namespace,
		"ip", "link",
		"set", "dev", n.VirtIface, "up",
	)
	if err != nil {
		return
	}

	return
}

func (n *NetConf) Space(db *database.Database) (err error) {
	err = n.spaceSysctl(db)
	if err != nil {
		return
	}

	err = n.spaceForward(db)
	if err != nil {
		return
	}

	err = n.spaceVirt(db)
	if err != nil {
		return
	}

	err = n.spaceLoopback(db)
	if err != nil {
		return
	}

	err = n.spaceMtu(db)
	if err != nil {
		return
	}

	err = n.spaceUp(db)
	if err != nil {
		return
	}

	return
}

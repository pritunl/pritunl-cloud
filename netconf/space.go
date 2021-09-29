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

	if n.NetworkMode6 != node.Disabled {
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "netns", "exec", n.Namespace,
			"sysctl", "-w",
			fmt.Sprintf("net.ipv6.conf.%s.accept_ra=2",
				n.SpaceExternalIface6),
		)
		if err != nil {
			return
		}
	}

	return
}

func (n *NetConf) spaceForward(db *database.Database) (err error) {
	if n.NetworkMode != node.Disabled {
		iptables.Lock()
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "netns", "exec", n.Namespace,
			"iptables",
			"-I", "FORWARD", "1",
			"!", "-d", n.InternalAddr.String()+"/32",
			"-i", n.SpaceExternalIface,
			"-j", "DROP",
		)
		iptables.Unlock()
		if err != nil {
			return
		}
	}

	if n.NetworkMode6 != node.Disabled {
		iptables.Lock()
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "netns", "exec", n.Namespace,
			"ip6tables",
			"-I", "FORWARD", "1",
			"!", "-d", n.InternalAddr6.String()+"/128",
			"-i", n.SpaceExternalIface6,
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

func (n *NetConf) Space(db *database.Database) (err error) {
	err = n.spaceSysctl(db)
	if err != nil {
		return
	}

	err = n.spaceForward(db)
	if err != nil {
		return
	}

	return
}

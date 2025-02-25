package netconf

import (
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/iproute"
	"github.com/pritunl/pritunl-cloud/iptables"
	"github.com/pritunl/pritunl-cloud/utils"
)

func (n *NetConf) bridgeNet(db *database.Database) (err error) {
	err = iproute.BridgeAdd(n.Namespace, n.SpaceBridgeIface)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", n.Namespace,
		"sysctl", "-w",
		"net.ipv4.conf.br0.arp_accept=0",
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", n.Namespace,
		"sysctl", "-w",
		"net.ipv4.conf.br0.arp_ignore=2",
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", n.Namespace,
		"sysctl", "-w",
		"net.ipv4.conf.br0.arp_filter=1",
	)
	if err != nil {
		return
	}

	return
}

func (n *NetConf) bridgeMaster(db *database.Database) (err error) {
	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", n.Namespace,
		"ip", "link", "set",
		n.BridgeInternalIface, "master", n.SpaceBridgeIface,
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", n.Namespace,
		"ip", "link", "set",
		n.VirtIface, "master", n.SpaceBridgeIface,
	)
	if err != nil {
		return
	}

	return
}

func (n *NetConf) bridgeRoute(db *database.Database) (err error) {
	_, err = utils.ExecCombinedOutputLogged(
		[]string{"File exists"},
		"ip", "netns", "exec", n.Namespace,
		"ip", "addr",
		"add", n.InternalGatewayAddrCidr,
		"dev", n.SpaceBridgeIface,
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		[]string{"File exists"},
		"ip", "netns", "exec", n.Namespace,
		"ip", "-6", "addr",
		"add", n.InternalGatewayAddr6.String()+"/64",
		"dev", n.SpaceBridgeIface,
	)
	if err != nil {
		return
	}

	return
}

func (n *NetConf) bridgeIptables(db *database.Database) (err error) {
	iptables.Lock()
	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", n.Namespace,
		"ebtables",
		"-A", "INPUT",
		"-p", "ARP",
		"-i", "!", n.VirtIface,
		"--arp-ip-dst", n.InternalGatewayAddr.String(),
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
		"ebtables",
		"-A", "OUTPUT",
		"-p", "ARP",
		"-o", "!", n.VirtIface,
		"--arp-ip-dst", n.InternalGatewayAddr.String(),
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
		"ebtables",
		"-A", "FORWARD",
		"-p", "ARP",
		"-o", "!", n.VirtIface,
		"--arp-ip-dst", n.InternalGatewayAddr.String(),
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
		"ebtables",
		"-A", "INPUT",
		"-p", "ARP",
		"-i", "!", n.VirtIface,
		"--arp-ip-src", n.InternalGatewayAddr.String(),
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
		"ebtables",
		"-A", "OUTPUT",
		"-p", "ARP",
		"-o", "!", n.VirtIface,
		"--arp-ip-src", n.InternalGatewayAddr.String(),
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
		"ebtables",
		"-A", "FORWARD",
		"-p", "ARP",
		"-o", "!", n.VirtIface,
		"--arp-ip-src", n.InternalGatewayAddr.String(),
		"-j", "DROP",
	)
	iptables.Unlock()
	if err != nil {
		return
	}

	return
}

func (n *NetConf) bridgeUp(db *database.Database) (err error) {
	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", n.Namespace,
		"ip", "link",
		"set", "dev", n.SpaceBridgeIface, "up",
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", n.Namespace,
		"bridge", "link",
		"set", "dev", n.VirtIface, "hairpin", "on",
	)
	if err != nil {
		return
	}

	return
}

func (n *NetConf) Bridge(db *database.Database) (err error) {
	err = n.bridgeNet(db)
	if err != nil {
		return
	}

	err = n.bridgeMaster(db)
	if err != nil {
		return
	}

	err = n.bridgeRoute(db)
	if err != nil {
		return
	}

	err = n.bridgeIptables(db)
	if err != nil {
		return
	}

	err = n.bridgeUp(db)
	if err != nil {
		return
	}

	return
}

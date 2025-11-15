package dhcps

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/datacenter"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/features"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/permission"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/systemd"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/pritunl/pritunl-cloud/vpc"
	"github.com/pritunl/pritunl-cloud/zone"
	"github.com/sirupsen/logrus"
)

const (
	dhcpCaps = "CAP_NET_BIND_SERVICE CAP_NET_BROADCAST"
	ndpCaps  = "CAP_NET_BIND_SERVICE CAP_NET_BROADCAST CAP_NET_RAW"
)

const systemdTemplate = `[Unit]
Description=Pritunl Cloud %s
After=network.target

[Service]
Environment=CONFIG='%s'
Type=simple
User=root
ExecStart=/usr/sbin/ip netns exec %s %s %s
TimeoutStopSec=5
PrivateTmp=true
ProtectHome=true
ProtectSystem=full
ProtectHostname=true
ProtectKernelTunables=true
AmbientCapabilities=%s
`

const systemdNamespaceTemplate = `[Unit]
Description=Pritunl Cloud %s
After=network.target

[Service]
Environment=CONFIG='%s'
Type=simple
User=%s
ExecStart=%s %s
TimeoutStopSec=5
PrivateTmp=true
ProtectHome=true
ProtectSystem=full
ProtectHostname=true
ProtectKernelTunables=true
NetworkNamespacePath=/var/run/netns/%s
AmbientCapabilities=%s
`

func UpdateEbtables(vmId bson.ObjectID, namespace string) (err error) {
	iface := vm.GetIface(vmId, 0)

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", namespace,
		"iptables",
		"-t", "mangle",
		"-A", "POSTROUTING",
		"-o", settings.Hypervisor.BridgeIfaceName,
		"-p", "udp",
		"-m", "udp",
		"--sport", "67",
		"-j", "CHECKSUM",
		"--checksum-fill",
	)
	if err != nil {
		return
	}
	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", namespace,
		"ip6tables",
		"-t", "mangle",
		"-A", "POSTROUTING",
		"-o", settings.Hypervisor.BridgeIfaceName,
		"-p", "udp",
		"-m", "udp",
		"--sport", "547",
		"-j", "CHECKSUM",
		"--checksum-fill",
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", namespace,
		"ebtables",
		"-I", "OUTPUT",
		"-o", iface,
		"-p", "IPv4",
		"--ip-protocol", "udp",
		"--ip-sport", "67",
		"-j", "ACCEPT",
	)
	if err != nil {
		return
	}
	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", namespace,
		"ebtables",
		"-A", "OUTPUT",
		"-p", "IPv4",
		"--ip-protocol", "udp",
		"--ip-sport", "67",
		"-j", "DROP",
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", namespace,
		"ebtables",
		"-I", "OUTPUT",
		"-o", iface,
		"-p", "IPv6",
		"--ip6-protocol", "udp",
		"--ip6-sport", "547",
		"-j", "ACCEPT",
	)
	if err != nil {
		return
	}
	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", namespace,
		"ebtables",
		"-A", "OUTPUT",
		"-p", "IPv6",
		"--ip6-protocol", "udp",
		"--ip6-sport", "547",
		"-j", "DROP",
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", namespace,
		"ebtables",
		"-I", "OUTPUT",
		"-o", iface,
		"-p", "IPv6",
		"--ip6-protocol", "ipv6-icmp",
		"--ip6-icmp-type", "134",
		"-j", "ACCEPT",
	)
	if err != nil {
		return
	}
	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", namespace,
		"ebtables",
		"-A", "OUTPUT",
		"-p", "IPv6",
		"--ip6-protocol", "ipv6-icmp",
		"--ip6-icmp-type", "134",
		"-j", "DROP",
	)
	if err != nil {
		return
	}

	return
}

func ClearEbtables(vmId bson.ObjectID, namespace string) (err error) {
	iface := vm.GetIface(vmId, 0)

	_, _ = utils.ExecCombinedOutput(
		"",
		"ip", "netns", "exec", namespace,
		"iptables",
		"-t", "mangle",
		"-D", "POSTROUTING",
		"-o", settings.Hypervisor.BridgeIfaceName,
		"-p", "udp",
		"-m", "udp",
		"--sport", "67",
		"-j", "CHECKSUM",
		"--checksum-fill",
	)
	_, _ = utils.ExecCombinedOutput(
		"",
		"ip", "netns", "exec", namespace,
		"ip6tables",
		"-t", "mangle",
		"-D", "POSTROUTING",
		"-o", settings.Hypervisor.BridgeIfaceName,
		"-p", "udp",
		"-m", "udp",
		"--sport", "547",
		"-j", "CHECKSUM",
		"--checksum-fill",
	)

	_, _ = utils.ExecCombinedOutput(
		"",
		"ip", "netns", "exec", namespace,
		"ebtables",
		"-D", "OUTPUT",
		"-o", iface,
		"-p", "IPv4",
		"--ip-protocol", "udp",
		"--ip-sport", "67",
		"-j", "ACCEPT",
	)
	_, _ = utils.ExecCombinedOutput(
		"",
		"ip", "netns", "exec", namespace,
		"ebtables",
		"-D", "OUTPUT",
		"-p", "IPv4",
		"--ip-protocol", "udp",
		"--ip-sport", "67",
		"-j", "DROP",
	)

	_, _ = utils.ExecCombinedOutput(
		"",
		"ip", "netns", "exec", namespace,
		"ebtables",
		"-D", "OUTPUT",
		"-o", iface,
		"-p", "IPv6",
		"--ip6-protocol", "udp",
		"--ip6-sport", "547",
		"-j", "ACCEPT",
	)
	_, _ = utils.ExecCombinedOutput(
		"",
		"ip", "netns", "exec", namespace,
		"ebtables",
		"-D", "OUTPUT",
		"-p", "IPv6",
		"--ip6-protocol", "udp",
		"--ip6-sport", "547",
		"-j", "DROP",
	)

	_, _ = utils.ExecCombinedOutput(
		"",
		"ip", "netns", "exec", namespace,
		"ebtables",
		"-D", "OUTPUT",
		"-o", iface,
		"-p", "IPv6",
		"--ip6-protocol", "ipv6-icmp",
		"--ip6-icmp-type", "134",
		"-j", "ACCEPT",
	)
	_, _ = utils.ExecCombinedOutput(
		"",
		"ip", "netns", "exec", namespace,
		"ebtables",
		"-D", "OUTPUT",
		"-p", "IPv6",
		"--ip6-protocol", "ipv6-icmp",
		"--ip6-icmp-type", "134",
		"-j", "DROP",
	)

	return
}

func WriteService(vmId bson.ObjectID, namespace string,
	config interface{}, systemdNamespace bool) (err error) {

	param := ""
	unitPath := ""
	caps := ""

	curPath, err := os.Executable()
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "dhcps: Failed to get executable path"),
		}
		return
	}

	confData, err := json.Marshal(config)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "dhcps: Failed to marshal config"),
		}
		return
	}

	switch config.(type) {
	case *Server4:
		param = "dhcp4-server"
		unitPath = paths.GetUnitPathDhcp4(vmId, 0)
		caps = dhcpCaps
		break
	case *Server6:
		param = "dhcp6-server"
		unitPath = paths.GetUnitPathDhcp6(vmId, 0)
		caps = dhcpCaps
		break
	case *ServerNdp:
		param = "ndp-server"
		unitPath = paths.GetUnitPathNdp(vmId, 0)
		caps = ndpCaps
		break
	default:
		err = &errortypes.TypeError{
			errors.New("dhcps: Unknown config type"),
		}
		return
	}

	output := ""
	if systemdNamespace {
		output = fmt.Sprintf(
			systemdNamespaceTemplate,
			param,
			string(confData),
			permission.GetUserName(vmId),
			curPath,
			param,
			namespace,
			caps,
		)
	} else {
		output = fmt.Sprintf(
			systemdTemplate,
			param,
			string(confData),
			namespace,
			curPath,
			param,
			caps,
		)
	}

	err = utils.CreateWrite(unitPath, output, 0644)
	if err != nil {
		return
	}

	return
}

func Start(db *database.Database, virt *vm.VirtualMachine,
	dc *datacenter.Datacenter, zne *zone.Zone, vc *vpc.Vpc) (err error) {

	namespace := vm.GetNamespace(virt.Id, 0)

	hasSystemdNamespace := features.HasSystemdNamespace()

	logrus.WithFields(logrus.Fields{
		"id": virt.Id.Hex(),
	}).Info("dhcps: Starting virtual machine dhcp server")

	if virt.NetworkAdapters == nil || len(virt.NetworkAdapters) < 1 {
		err = &errortypes.ParseError{
			errors.New("dhcps: Missing virt network adapter"),
		}
		return
	}
	subnetId := virt.NetworkAdapters[0].Subnet

	vcNet, err := vc.GetNetwork()
	if err != nil {
		return
	}

	cidr, _ := vcNet.Mask.Size()
	addr, gatewayAddr, err := vc.GetIp(db, subnetId, virt.Id)
	if err != nil {
		return
	}

	addr6 := vc.GetIp6(virt.Id)
	gatewayAddr6 := vc.GetGatewayIp6(virt.Id)

	mtu := dc.GetInstanceMtu()

	server4 := &Server4{
		Iface:     settings.Hypervisor.BridgeIfaceName,
		ClientIp:  addr.String(),
		GatewayIp: gatewayAddr.String(),
		PrefixLen: cidr,
		DnsServers: []string{
			strings.Split(settings.Hypervisor.ImdsAddress, "/")[0],
			zne.GetDnsServerPrimary(),
			zne.GetDnsServerSecondary(),
		},
		Mtu:      mtu,
		Lifetime: settings.Hypervisor.DhcpLifetime,
	}
	server6 := &Server6{
		Iface:      settings.Hypervisor.BridgeIfaceName,
		ClientIp:   addr6.String(),
		GatewayIp:  gatewayAddr6.String(),
		PrefixLen:  64,
		DnsServers: []string{},
		Mtu:        mtu,
		Lifetime:   settings.Hypervisor.DhcpLifetime,
	}
	serverNdp := &ServerNdp{
		Iface:      settings.Hypervisor.BridgeIfaceName,
		ClientIp:   addr6.String(),
		GatewayIp:  gatewayAddr6.String(),
		PrefixLen:  64,
		DnsServers: []string{},
		Mtu:        mtu,
		Lifetime:   settings.Hypervisor.DhcpLifetime,
		Delay:      settings.Hypervisor.NdpRaInterval,
	}

	err = UpdateEbtables(virt.Id, namespace)
	if err != nil {
		return
	}

	unitServer4 := paths.GetUnitNameDhcp4(virt.Id, 0)
	unitServer6 := paths.GetUnitNameDhcp6(virt.Id, 0)
	unitServerNdp := paths.GetUnitNameNdp(virt.Id, 0)

	_ = systemd.Stop(unitServer4)
	_ = systemd.Stop(unitServer6)
	_ = systemd.Stop(unitServerNdp)

	err = WriteService(virt.Id, namespace, server4, hasSystemdNamespace)
	if err != nil {
		return
	}
	err = WriteService(virt.Id, namespace, server6, hasSystemdNamespace)
	if err != nil {
		return
	}
	err = WriteService(virt.Id, namespace, serverNdp, hasSystemdNamespace)
	if err != nil {
		return
	}

	err = systemd.Reload()
	if err != nil {
		return
	}

	err = systemd.Start(unitServer4)
	if err != nil {
		return
	}
	err = systemd.Start(unitServer6)
	if err != nil {
		return
	}
	err = systemd.Start(unitServerNdp)
	if err != nil {
		return
	}

	return
}

func Stop(virt *vm.VirtualMachine) (err error) {
	namespace := vm.GetNamespace(virt.Id, 0)
	unitServer4 := paths.GetUnitNameDhcp4(virt.Id, 0)
	unitServer6 := paths.GetUnitNameDhcp6(virt.Id, 0)
	unitServerNdp := paths.GetUnitNameNdp(virt.Id, 0)

	_ = systemd.Stop(unitServer4)
	_ = systemd.Stop(unitServer6)
	_ = systemd.Stop(unitServerNdp)

	err = ClearEbtables(virt.Id, namespace)
	if err != nil {
		return
	}

	return
}

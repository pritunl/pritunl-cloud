package dhcpc

import (
	"fmt"
	"strings"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/features"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/permission"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/systemd"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/sirupsen/logrus"
)

const systemdNamespaceTemplate = `[Unit]
Description=Pritunl Cloud DHCP Client
After=network.target

[Service]
Type=simple
User=%s
Environment="IMDS_ADDRESS=%s"
Environment="IMDS_PORT=%d"
Environment="IMDS_SECRET=%s"
Environment="DHCP_IFACE=%s"
Environment="DHCP_IFACE6=%s"
Environment="DHCP_IP=%s"
Environment="DHCP_IP6=%s"
ExecStart=/usr/bin/pritunl-cloud %s dhcp-client
TimeoutStopSec=5
Restart=always
RestartSec=3
PrivateTmp=true
ProtectHome=true
ProtectSystem=full
ProtectHostname=true
ProtectKernelTunables=true
NetworkNamespacePath=/var/run/netns/%s
AmbientCapabilities=CAP_NET_RAW CAP_NET_BIND_SERVICE CAP_NET_ADMIN
`

const systemdTemplate = `[Unit]
Description=Pritunl Cloud DHCP Client
After=network.target

[Service]
Type=simple
User=root
Environment="IMDS_ADDRESS=%s"
Environment="IMDS_PORT=%d"
Environment="IMDS_SECRET=%s"
Environment="DHCP_IFACE=%s"
Environment="DHCP_IFACE6=%s"
Environment="DHCP_IP=%s"
Environment="DHCP_IP6=%s"
ExecStart=/usr/sbin/ip netns exec %s /usr/bin/pritunl-cloud %s dhcp-client
TimeoutStopSec=5
Restart=always
RestartSec=3
PrivateTmp=true
ProtectHome=true
ProtectSystem=full
ProtectHostname=true
ProtectKernelTunables=true
AmbientCapabilities=CAP_NET_RAW CAP_NET_BIND_SERVICE CAP_NET_ADMIN
`

func WriteService(vmId primitive.ObjectID,
	namespace, imdsSecret, dhcpIface, dhcpIface6, dhcpIp, dhcpIp6 string,
	ip4, ip6, systemdNamespace bool) (err error) {

	unitPath := paths.GetUnitPathDhcpc(vmId)

	if imdsSecret == "" {
		err = &errortypes.ParseError{
			errors.New("dhcpc: Cannot start dhcp client with empty secret"),
		}
		return
	}

	if dhcpIface == "" {
		err = &errortypes.ParseError{
			errors.New("dhcpc: Cannot start dhcp client with empty iface"),
		}
		return
	}

	args := []string{}
	if ip4 {
		args = append(args, "-ip4")
	}
	if ip6 {
		args = append(args, "-ip6")
	}

	output := ""
	if systemdNamespace {
		output = fmt.Sprintf(
			systemdNamespaceTemplate,
			permission.GetUserName(vmId),
			settings.Hypervisor.ImdsAddress,
			settings.Hypervisor.ImdsPort,
			imdsSecret,
			dhcpIface,
			dhcpIface6,
			dhcpIp,
			dhcpIp6,
			strings.Join(args, " "),
			namespace,
		)
	} else {
		output = fmt.Sprintf(
			systemdTemplate,
			settings.Hypervisor.ImdsAddress,
			settings.Hypervisor.ImdsPort,
			imdsSecret,
			dhcpIface,
			dhcpIface6,
			dhcpIp,
			dhcpIp6,
			strings.Join(args, " "),
			namespace,
		)
	}

	err = utils.CreateWrite(unitPath, output, 0600)
	if err != nil {
		return
	}

	return
}

func Start(db *database.Database, virt *vm.VirtualMachine,
	iface, iface6 string, ip4, ip6 bool) (err error) {

	namespace := vm.GetNamespace(virt.Id, 0)

	hasSystemdNamespace := features.HasSystemdNamespace()
	unit := paths.GetUnitNameDhcpc(virt.Id)

	logrus.WithFields(logrus.Fields{
		"id":           virt.Id.Hex(),
		"systemd_unit": unit,
	}).Info("dhcpc: Starting virtual machine dhcp client")

	_ = systemd.Stop(unit)

	err = WriteService(virt.Id, namespace, virt.ImdsDhcpSecret, iface, iface6,
		virt.DhcpIp, virt.DhcpIp6, ip4, ip6, hasSystemdNamespace)
	if err != nil {
		return
	}

	err = systemd.Reload()
	if err != nil {
		return
	}

	err = systemd.Start(unit)
	if err != nil {
		return
	}

	return
}

func Stop(virt *vm.VirtualMachine) (err error) {
	unit := paths.GetUnitNameDhcpc(virt.Id)

	_ = systemd.Stop(unit)

	return
}

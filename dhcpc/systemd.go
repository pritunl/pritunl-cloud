package dhcpc

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
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

[Service]%s
Type=simple
User=%s
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

[Service]%s
Type=simple
User=root
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

func WriteService(vmId bson.ObjectID,
	namespace, imdsSecret, dhcpIface, dhcpIface6, dhcpIp, dhcpIp6 string,
	ip4, ip6, systemdNamespace bool) (err error) {

	unitPath := paths.GetUnitPathDhcpc(vmId)
	imdsAddress := utils.FilterIp(
		strings.Split(settings.Hypervisor.ImdsAddress, "/")[0])
	namespace = utils.FilterName(namespace)
	dhcpIface = utils.FilterIface(dhcpIface)
	dhcpIface6 = utils.FilterIface(dhcpIface6)

	if imdsAddress == "" {
		err = &errortypes.ParseError{
			errors.New("dhcpc: Cannot start dhcp client with invalid address"),
		}
		return
	}

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

	if namespace == "" {
		err = &errortypes.ParseError{
			errors.New("dhcpc: Cannot start dhcp client with empty namespace"),
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

	env, err := utils.JoinSystemdEnv(
		utils.FilterSystemdEnv("IMDS_ADDRESS", imdsAddress),
		utils.FilterSystemdEnv("IMDS_PORT",
			strconv.Itoa(settings.Hypervisor.ImdsPort)),
		utils.FilterSystemdEnv("IMDS_SECRET", imdsSecret),
		utils.FilterSystemdEnv("DHCP_IFACE", dhcpIface),
		utils.FilterSystemdEnv("DHCP_IFACE6", dhcpIface6),
		utils.FilterSystemdEnv("DHCP_IP", utils.FilterCidr(dhcpIp)),
		utils.FilterSystemdEnv("DHCP_IP6", utils.FilterCidr(dhcpIp6)),
		utils.FilterSystemdEnv("DHCP_INTERVAL",
			strconv.Itoa(settings.Hypervisor.DhcpRenewTtl)),
	)
	if err != nil {
		return
	}

	output := ""
	if systemdNamespace {
		output = fmt.Sprintf(
			systemdNamespaceTemplate,
			env,
			permission.GetUserName(vmId),
			strings.Join(args, " "),
			namespace,
		)
	} else {
		output = fmt.Sprintf(
			systemdTemplate,
			env,
			namespace,
			strings.Join(args, " "),
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
	unitPath := paths.GetUnitPathDhcpc(virt.Id)

	_ = systemd.Stop(unit)

	_ = utils.RemoveAll(unitPath)

	return
}

func Restart(vmId bson.ObjectID) (err error) {
	unit := paths.GetUnitNameDhcpc(vmId)

	err = systemd.Restart(unit)
	return
}

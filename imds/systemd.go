package imds

import (
	"fmt"

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
Description=Pritunl Cloud IMDS Server
After=network.target

[Service]
Type=simple
User=%s
Environment="CLIENT_SECRET=%s"
Environment="DHCP_SECRET=%s"
Environment="HOST_SECRET=%s"
ExecStart=/usr/bin/pritunl-cloud -sock=%s -host=%s -port=%d imds-server
TimeoutStopSec=5
PrivateTmp=true
ProtectHome=true
ProtectSystem=full
ProtectHostname=true
ProtectKernelTunables=true
NetworkNamespacePath=/var/run/netns/%s
AmbientCapabilities=CAP_NET_BIND_SERVICE
`

const systemdTemplate = `[Unit]
Description=Pritunl Cloud IMDS Server
After=network.target

[Service]
Type=simple
User=root
Environment="CLIENT_SECRET=%s"
Environment="DHCP_SECRET=%s"
Environment="HOST_SECRET=%s"
ExecStart=/usr/sbin/ip netns exec %s /usr/bin/pritunl-cloud -sock=%s -host=%s -port=%d imds-server
TimeoutStopSec=5
PrivateTmp=true
ProtectHome=true
ProtectSystem=full
ProtectHostname=true
ProtectKernelTunables=true
AmbientCapabilities=CAP_NET_BIND_SERVICE
`

func WriteService(vmId primitive.ObjectID,
	namespace, clientSecret, dhcpSecret, hostSecret string,
	systemdNamespace bool) (err error) {

	unitPath := paths.GetUnitPathImds(vmId)
	sockPath := paths.GetImdsSockPath(vmId)

	if clientSecret == "" || dhcpSecret == "" || hostSecret == "" {
		err = &errortypes.ParseError{
			errors.New("imds: Cannot start imds with empty secret"),
		}
		return
	}

	output := ""
	if systemdNamespace {
		output = fmt.Sprintf(
			systemdNamespaceTemplate,
			permission.GetUserName(vmId),
			clientSecret,
			dhcpSecret,
			hostSecret,
			sockPath,
			settings.Hypervisor.ImdsAddress,
			settings.Hypervisor.ImdsPort,
			namespace,
		)
	} else {
		output = fmt.Sprintf(
			systemdTemplate,
			clientSecret,
			dhcpSecret,
			hostSecret,
			namespace,
			sockPath,
			settings.Hypervisor.ImdsAddress,
			settings.Hypervisor.ImdsPort,
		)
	}

	err = utils.CreateWrite(unitPath, output, 0600)
	if err != nil {
		return
	}

	return
}

func Start(db *database.Database, virt *vm.VirtualMachine) (err error) {
	namespace := vm.GetNamespace(virt.Id, 0)

	hasSystemdNamespace := features.HasSystemdNamespace()
	unit := paths.GetUnitNameImds(virt.Id)

	logrus.WithFields(logrus.Fields{
		"id":           virt.Id.Hex(),
		"systemd_unit": unit,
	}).Info("imds: Starting virtual machine imds server")

	_ = systemd.Stop(unit)

	err = WriteService(virt.Id, namespace, virt.ImdsClientSecret,
		virt.ImdsDhcpSecret, virt.ImdsHostSecret, hasSystemdNamespace)
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
	unit := paths.GetUnitNameImds(virt.Id)

	_ = systemd.Stop(unit)

	return
}

package imds

import (
	"fmt"
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
Description=Pritunl Cloud IMDS Server
After=network.target

[Service]%s
Type=simple
User=%s
ExecStart=/usr/bin/pritunl-cloud -sock=%s -host=%s -port=%d imds-server
TimeoutStopSec=5
Restart=always
RestartSec=3
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

[Service]%s
Type=simple
User=root
ExecStart=/usr/sbin/ip netns exec %s /usr/bin/pritunl-cloud -sock=%s -host=%s -port=%d imds-server
TimeoutStopSec=5
Restart=always
RestartSec=3
PrivateTmp=true
ProtectHome=true
ProtectSystem=full
ProtectHostname=true
ProtectKernelTunables=true
AmbientCapabilities=CAP_NET_BIND_SERVICE
`

func WriteService(vmId bson.ObjectID,
	namespace, clientSecret, dhcpSecret, hostSecret string,
	systemdNamespace bool) (err error) {

	unitPath := paths.GetUnitPathImds(vmId)
	sockPath := utils.FilterPath(paths.GetImdsSockPath(vmId))
	imdsAddress := utils.FilterIp(
		strings.Split(settings.Hypervisor.ImdsAddress, "/")[0])
	namespace = utils.FilterName(namespace)

	if clientSecret == "" || dhcpSecret == "" || hostSecret == "" {
		err = &errortypes.ParseError{
			errors.New("imds: Cannot start imds with empty secret"),
		}
		return
	}

	if imdsAddress == "" {
		err = &errortypes.ParseError{
			errors.New("imds: Cannot start imds with invalid address"),
		}
		return
	}

	if namespace == "" {
		err = &errortypes.ParseError{
			errors.New("imds: Cannot start imds with empty namespace"),
		}
		return
	}

	env := ""
	env += "\n" + utils.FilterSystemdEnv("INSTANCE_ID", vmId.Hex())
	env += "\n" + utils.FilterSystemdEnv("CLIENT_SECRET", clientSecret)
	env += "\n" + utils.FilterSystemdEnv("DHCP_SECRET", dhcpSecret)
	env += "\n" + utils.FilterSystemdEnv("HOST_SECRET", hostSecret)

	output := ""
	if systemdNamespace {
		output = fmt.Sprintf(
			systemdNamespaceTemplate,
			env,
			permission.GetUserName(vmId),
			sockPath,
			imdsAddress,
			settings.Hypervisor.ImdsPort,
			namespace,
		)
	} else {
		output = fmt.Sprintf(
			systemdTemplate,
			env,
			namespace,
			sockPath,
			imdsAddress,
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

func Restart(instId bson.ObjectID) (err error) {
	unit := paths.GetUnitNameImds(instId)

	_ = systemd.Restart(unit)

	return
}

func Stop(virt *vm.VirtualMachine) (err error) {
	unit := paths.GetUnitNameImds(virt.Id)

	_ = systemd.Stop(unit)

	return
}

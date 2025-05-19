package virtiofs

import (
	"fmt"

	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/systemd"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/sirupsen/logrus"
)

const systemdTemplate = `[Unit]
Description=Pritunl Cloud VirtIO-FS Daemon
After=network.target

[Service]
Type=simple
User=root
ExecStart=/usr/libexec/virtiofsd --socket-path="%s" --shared-dir="%s" --sandbox=namespace
TimeoutStopSec=5
PrivateTmp=true
ProtectSystem=full
ProtectHostname=true
ProtectKernelTunables=true
`

func WriteService(vmId primitive.ObjectID,
	shareId, sharePath string) (err error) {

	unitPath := paths.GetUnitPathShare(vmId, shareId)
	sockPath := paths.GetShareSockPath(vmId, shareId)

	output := fmt.Sprintf(
		systemdTemplate,
		sockPath,
		sharePath,
	)

	err = utils.CreateWrite(unitPath, output, 0600)
	if err != nil {
		return
	}

	return
}

func Start(db *database.Database, virt *vm.VirtualMachine,
	shareId, sharePath string) (err error) {

	unit := paths.GetUnitNameShare(virt.Id, shareId)

	logrus.WithFields(logrus.Fields{
		"id":           virt.Id.Hex(),
		"systemd_unit": unit,
	}).Info("virtiofs: Starting virtual machine virtiofsd")

	_ = systemd.Stop(unit)

	err = WriteService(virt.Id, shareId, sharePath)
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

func Stop(virt *vm.VirtualMachine, shareId string) (err error) {
	unit := paths.GetUnitNameShare(virt.Id, shareId)

	_ = systemd.Stop(unit)

	return
}

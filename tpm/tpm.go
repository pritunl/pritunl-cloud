package tpm

import (
	"fmt"
	"os"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/permission"
	"github.com/pritunl/pritunl-cloud/systemd"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/sirupsen/logrus"
)

const systemdTemplate = `[Unit]
Description=Pritunl Cloud TPM
After=network.target

[Service]
Type=simple
User=%s
ExecStart=swtpm socket --tpm2 --key pwdfile=%s,mode=aes-256-cbc,remove=true,kdf=pbkdf2 --tpmstate dir=%s --ctrl type=unixio,path=%s --log level=5
TimeoutStopSec=5
PrivateTmp=true
ProtectHome=true
ProtectSystem=full
ProtectHostname=true
ProtectKernelTunables=true
NetworkNamespacePath=/var/run/netns/%s
`

func WriteService(vmId primitive.ObjectID, namespace string) (err error) {
	unitPath := paths.GetUnitPathTpm(vmId)
	tpmPath := paths.GetTpmPath(vmId)
	pwdPath := paths.GetTpmPwdPath(vmId)
	sockPath := paths.GetTpmSockPath(vmId)

	output := fmt.Sprintf(
		systemdTemplate,
		permission.GetUserName(vmId),
		pwdPath,
		tpmPath,
		sockPath,
		namespace,
	)

	err = utils.CreateWrite(unitPath, output, 0644)
	if err != nil {
		return
	}

	return
}

func Start(db *database.Database, virt *vm.VirtualMachine) (err error) {
	namespace := vm.GetNamespace(virt.Id, 0)

	logrus.WithFields(logrus.Fields{
		"id": virt.Id.Hex(),
	}).Info("tpm: Starting virtual machine tpm")

	tpmsPath := paths.GetTpmsPath()
	tpmPath := paths.GetTpmPath(virt.Id)
	unit := paths.GetUnitNameTpm(virt.Id)
	pwdPath := paths.GetTpmPwdPath(virt.Id)

	err = utils.ExistsMkdir(tpmsPath, 0755)
	if err != nil {
		return
	}

	err = utils.ExistsMkdir(tpmPath, 0700)
	if err != nil {
		return
	}

	err = permission.InitTpm(virt)
	if err != nil {
		return
	}

	_ = systemd.Stop(unit)

	secret, err := GetSecret(db, virt.Id)
	if err != nil {
		return
	}

	if secret == "" {
		err = &errortypes.NotFoundError{
			errors.New("tpm: Missing instance tpm secret"),
		}
		return
	}

	err = WriteService(virt.Id, namespace)
	if err != nil {
		return
	}

	err = systemd.Reload()
	if err != nil {
		return
	}

	go func() {
		time.Sleep(15 * time.Second)
		_ = os.Remove(pwdPath)
	}()

	err = utils.CreateWrite(pwdPath, secret, 0600)
	if err != nil {
		return
	}

	err = permission.InitTpmPwd(virt)
	if err != nil {
		_ = os.Remove(pwdPath)
		return
	}

	err = systemd.Start(unit)
	if err != nil {
		_ = os.Remove(pwdPath)
		return
	}

	return
}

func Stop(virt *vm.VirtualMachine) (err error) {
	unit := paths.GetUnitNameTpm(virt.Id)

	_ = systemd.Stop(unit)

	return
}

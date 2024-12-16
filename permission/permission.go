package permission

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
)

func chown(virt *vm.VirtualMachine, path string) (err error) {
	err = os.Chown(path, virt.UnixId, 0)
	if err != nil {
		err = &errortypes.WriteError{
			errors.Newf(
				"permission: Failed to set owner of '%s' to '%d'",
				path, virt.UnixId,
			),
		}
		return
	}

	return
}

func touchChown(virt *vm.VirtualMachine, path string) (err error) {
	_, err = utils.ExecCombinedOutputLogged(nil,
		"touch", path,
	)
	if err != nil {
		return
	}

	err = chown(virt, path)
	if err != nil {
		return
	}

	return
}

func mkdirChown(virt *vm.VirtualMachine, path string) (err error) {
	_, err = utils.ExecCombinedOutputLogged(nil,
		"mkdir", path,
	)
	if err != nil {
		return
	}

	err = chown(virt, path)
	if err != nil {
		return
	}

	return
}

func InitVirt(virt *vm.VirtualMachine) (err error) {
	err = UserAdd(virt)
	if err != nil {
		return
	}

	if virt.Uefi {
		err = chown(virt, paths.GetOvmfVarsPath(virt.Id))
		if err != nil {
			return
		}
	}

	err = chown(virt, paths.GetInitPath(virt.Id))
	if err != nil {
		return
	}

	for _, disk := range virt.Disks {
		err = chown(virt, disk.Path)
		if err != nil {
			return
		}
	}

	for _, device := range virt.DriveDevices {
		drivePth := ""
		if device.Type == vm.Lvm {
			drivePth = filepath.Join("/dev/mapper",
				fmt.Sprintf("%s-%s", device.VgName, device.LvName))
		} else {
			drivePth = paths.GetDrivePath(device.Id)
		}

		err = chown(virt, drivePth)
		if err != nil {
			return
		}
	}

	err = chown(virt, paths.GetCacheDir(virt.Id))
	if err != nil {
		return
	}

	return
}

func InitDisk(virt *vm.VirtualMachine, dsk *vm.Disk) (err error) {
	err = UserAdd(virt)
	if err != nil {
		return
	}

	err = chown(virt, dsk.Path)
	if err != nil {
		return
	}

	return
}

func InitTpm(virt *vm.VirtualMachine) (err error) {
	tpmPath := paths.GetTpmPath(virt.Id)

	err = chown(virt, tpmPath)
	if err != nil {
		return
	}

	return
}

func InitTpmPwd(virt *vm.VirtualMachine) (err error) {
	tpmPath := paths.GetTpmPwdPath(virt.Id)

	err = chown(virt, tpmPath)
	if err != nil {
		return
	}

	return
}

func InitImds(virt *vm.VirtualMachine) (err error) {
	runPath := paths.GetInstRunPath(virt.Id)
	err = mkdirChown(virt, runPath)
	if err != nil {
		return
	}

	return
}

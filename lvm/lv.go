package lvm

import (
	"fmt"
	"github.com/pritunl/pritunl-cloud/utils"
	"path/filepath"
)

func CreateLv(vgName, lvName string, size int) (err error) {
	_, err = utils.ExecCombinedOutputLogged(nil,
		"lvcreate", "-an", "-L", fmt.Sprintf("%d.1G", size),
		"-n", lvName, vgName)
	if err != nil {
		return
	}

	return
}

func ActivateLv(vgName, lvName string) (err error) {
	_, err = utils.ExecCombinedOutputLogged(nil,
		"lvchange", "-ay", fmt.Sprintf("%s/%s", vgName, lvName))
	if err != nil {
		return
	}

	return
}

func WriteLv(vgName, lvName, sourcePth string) (err error) {
	dstPth := filepath.Join("/dev/mapper",
		fmt.Sprintf("%s-%s", vgName, lvName))

	_, err = utils.ExecCombinedOutputLogged(nil,
		"qemu-img", "convert", "-f", "qcow2",
		"-O", "raw", sourcePth, dstPth)
	if err != nil {
		return
	}

	return
}

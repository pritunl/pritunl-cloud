package lvm

import (
	"fmt"
	"math"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/utils"
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

func RemoveLv(vgName, lvName string) (err error) {
	_, err = utils.ExecCombinedOutputLogged([]string{
		"Failed to find",
	}, "lvremove", "-y", fmt.Sprintf("%s/%s", vgName, lvName))
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

func DeactivateLv(vgName, lvName string) (err error) {
	_, err = utils.ExecCombinedOutputLogged(nil,
		"lvchange", "-an", fmt.Sprintf("%s/%s", vgName, lvName))
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

func ExtendLv(vgName, lvName string, addSize int) (err error) {
	_, err = utils.ExecCombinedOutputLogged(nil,
		"lvextend", "-L", fmt.Sprintf("+%dG", addSize),
		fmt.Sprintf("%s/%s", vgName, lvName))
	if err != nil {
		return
	}

	return
}

func GetSizeLv(vgName, lvName string) (size int, err error) {
	output, err := utils.ExecCombinedOutput("",
		"lvs", fmt.Sprintf("%s/%s", vgName, lvName),
		"-o", "LV_SIZE", "--units", "g", "--noheadings")
	if err != nil {
		return
	}

	output = strings.Trim(strings.TrimSpace(strings.ToLower(output)), "g")

	number, err := strconv.ParseFloat(output, 64)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "lvm: Failed to parse lvm volume size"),
		}
		return
	}

	size = int(math.Round(number))

	return
}

func HasLocking(vgName string) (hasLock bool, err error) {
	output, err := utils.ExecCombinedOutput("",
		"vgs", vgName, "-o", "vg_lock_type", "--noheadings")
	if err != nil {
		return
	}

	lockType := strings.TrimSpace(output)
	hasLock = lockType != "" && lockType != "none"

	return
}

func IsLockspaceActive(vgName string) (isLocked bool, err error) {
	output, err := utils.ExecCombinedOutput("",
		"lvmlockctl", "-i")
	if err != nil {
		return
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, vgName) &&
			strings.Contains(line, "sanlock") {

			isLocked = true
			return
		}
	}

	return
}

func InitLock(vgName string) (err error) {
	hasLock, err := HasLocking(vgName)
	if err != nil {
		return
	}

	if !hasLock {
		return
	}

	isLocked, err := IsLockspaceActive(vgName)
	if err != nil {
		return
	}

	if isLocked {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(nil,
		"vgchange", "--lock-start", vgName)
	if err != nil {
		return
	}

	return
}

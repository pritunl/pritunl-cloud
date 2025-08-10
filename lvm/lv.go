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

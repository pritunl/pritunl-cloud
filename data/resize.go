package data

import (
	"encoding/json"
	"fmt"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/lock"
	"github.com/pritunl/pritunl-cloud/lvm"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/pool"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/sirupsen/logrus"
)

type diskInfo struct {
	Filename    string `json:"filename"`
	Format      string `json:"format"`
	ActualSize  int    `json:"actual-size"`
	VirtualSize int    `json:"virtual-size"`
}

func getDiskSizeQcow(dsk *disk.Disk) (size int, err error) {
	dskPth := paths.GetDiskPath(dsk.Id)

	output, err := utils.ExecOutput("",
		"qemu-img", "info", "--output=json", dskPth)
	if err != nil {
		return
	}

	diskInfo := &diskInfo{}

	err = json.Unmarshal([]byte(output), diskInfo)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "data: Failed to parse qemu disk info"),
		}
		return
	}

	size = diskInfo.VirtualSize / 1073741824

	return
}

func expandDiskQcow(db *database.Database, dsk *disk.Disk) (err error) {
	dskPth := paths.GetDiskPath(dsk.Id)

	logrus.WithFields(logrus.Fields{
		"disk_id":   dsk.Id.Hex(),
		"disk_path": dskPth,
		"new_size":  dsk.NewSize,
	}).Info("data: Expanding qcow disk")

	curSize, err := getDiskSizeQcow(dsk)
	if err != nil {
		return
	}

	if curSize >= dsk.NewSize {
		logrus.WithFields(logrus.Fields{
			"disk_id":      dsk.Id.Hex(),
			"disk_path":    dskPth,
			"current_size": curSize,
			"new_size":     dsk.NewSize,
		}).Warn("data: Disk size larger then new size")

		dsk.Size = curSize
		return
	}

	expandSize := dsk.NewSize - curSize
	_, err = utils.ExecCombinedOutputLogged(nil,
		"qemu-img", "resize", dskPth, fmt.Sprintf("+%dG", expandSize))
	if err != nil {
		return
	}

	curSize, err = getDiskSizeQcow(dsk)
	if err != nil {
		return
	}
	dsk.Size = curSize

	return
}

func expandDiskLvm(db *database.Database, dsk *disk.Disk) (err error) {
	pl, err := pool.Get(db, dsk.Pool)
	if err != nil {
		return
	}

	vgName := pl.VgName
	lvName := dsk.Id.Hex()

	logrus.WithFields(logrus.Fields{
		"disk_id":  dsk.Id.Hex(),
		"vg_name":  vgName,
		"lv_name":  lvName,
		"new_size": dsk.NewSize,
	}).Info("data: Expanding lvm disk")

	curSize, err := lvm.GetSizeLv(vgName, lvName)
	if err != nil {
		return
	}

	if curSize >= dsk.NewSize {
		logrus.WithFields(logrus.Fields{
			"disk_id":      dsk.Id.Hex(),
			"vg_name":      vgName,
			"lv_name":      lvName,
			"current_size": curSize,
			"new_size":     dsk.NewSize,
		}).Warn("data: Disk size larger then new size")

		dsk.Size = curSize
		return
	}

	acquired, err := lock.LvmLock(db, vgName, lvName)
	if err != nil {
		return
	}

	if !acquired {
		err = &errortypes.WriteError{
			errors.New("data: Failed to acquire LVM lock"),
		}
		return
	}
	defer func() {
		err2 := lock.LvmUnlock(db, vgName, lvName)
		if err2 != nil {
			logrus.WithFields(logrus.Fields{
				"error": err2,
			}).Error("data: Failed to unlock lvm")
		}
	}()

	err = lvm.ActivateLv(vgName, lvName)
	if err != nil {
		return
	}

	defer func() {
		err = lvm.DeactivateLv(vgName, lvName)
		if err != nil {
			return
		}
	}()

	expandSize := dsk.NewSize - curSize
	err = lvm.ExtendLv(vgName, lvName, expandSize)
	if err != nil {
		return
	}

	curSize, err = lvm.GetSizeLv(vgName, lvName)
	if err != nil {
		return
	}

	dsk.Size = curSize

	return
}

func ExpandDisk(db *database.Database, dsk *disk.Disk) (err error) {
	switch dsk.Type {
	case disk.Lvm:
		err = expandDiskLvm(db, dsk)
		if err != nil {
			return
		}
		break
	case "", disk.Qcow2:
		err = expandDiskQcow(db, dsk)
		if err != nil {
			return
		}
		break
	default:
		err = &errortypes.ParseError{
			errors.Newf("data: Unknown disk type %s", dsk.Type),
		}
		return
	}

	return
}

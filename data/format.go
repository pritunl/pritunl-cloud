package data

import (
	"fmt"
	"path/filepath"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/guestfs"
	"github.com/pritunl/pritunl-cloud/lock"
	"github.com/pritunl/pritunl-cloud/lvm"
	"github.com/pritunl/pritunl-cloud/pool"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/sirupsen/logrus"
)

const (
	mib = int64(1024 * 1024)
	gib = int64(1024 * 1024 * 1024)
)

func parseFileSystem(fileSystem string) (fsType string, fsLvm bool,
	err error) {

	switch fileSystem {
	case disk.Xfs:
		fsType = "xfs"
	case disk.LvmXfs:
		fsType = "xfs"
		fsLvm = true
	case disk.Ext4:
		fsType = "ext4"
	case disk.LvmExt4:
		fsType = "ext4"
		fsLvm = true
	default:
		err = &errortypes.WriteError{
			errors.Newf("data: Invalid disk filesystem %s", fileSystem),
		}
		return
	}

	return
}

func lvDevicePath(vgName, lvName string) string {
	return fmt.Sprintf("/dev/%s/%s", vgName, lvName)
}

func poolLvPath(vgName, lvName string) string {
	return filepath.Join("/dev/mapper",
		fmt.Sprintf("%s-%s", vgName, lvName))
}

func formatDevice(sess *guestfs.Session, dsk *disk.Disk) (
	uuid string, info *disk.Metadata, err error) {

	fsType, fsLvm, err := parseFileSystem(dsk.FileSystem)
	if err != nil {
		return
	}

	dev := sess.Device()

	sectorSize, err := sess.SectorSize(dev)
	if err != nil {
		return
	}

	alignSectors := disk.ResizeAlign / int64(sectorSize)
	if alignSectors < 1 {
		alignSectors = 1
	}

	err = sess.PartInit(dev, guestfs.Gpt)
	if err != nil {
		return
	}

	err = sess.PartAdd(dev, alignSectors, -(alignSectors + 1))
	if err != nil {
		return
	}

	partitions, err := sess.ListPartitions()
	if err != nil {
		return
	}

	if len(partitions) != 1 {
		err = &errortypes.WriteError{
			errors.Newf("data: Expected one partition found %d",
				len(partitions)),
		}
		return
	}
	partDev := partitions[0]

	fsDev := partDev
	if fsLvm {
		vgName := GetVgName(dsk.Id, 0)
		lvName := GetLvName(dsk.Id, 0)

		err = sess.PvCreate(partDev)
		if err != nil {
			return
		}

		err = sess.VgCreate(vgName, partDev)
		if err != nil {
			return
		}

		if dsk.LvSize == dsk.Size || dsk.LvSize <= 0 {
			err = sess.LvCreateFree(lvName, vgName, 100)
			if err != nil {
				return
			}
		} else {
			err = sess.LvCreate(lvName, vgName, int64(dsk.LvSize)*1024)
			if err != nil {
				return
			}
		}

		fsDev = lvDevicePath(vgName, lvName)
	}

	err = sess.Mkfs(fsType, fsDev)
	if err != nil {
		return
	}

	uuid, err = sess.VfsUuid(fsDev)
	if err != nil {
		return
	}

	if uuid == "" {
		err = &errortypes.ReadError{
			errors.New("data: Missing file system uuid after format"),
		}
		return
	}

	info, err = scanDevice(sess)
	if err != nil {
		return
	}

	return
}

func commitFormat(db *database.Database, dsk *disk.Disk,
	uuid string, info *disk.Metadata) (err error) {

	dsk.Uuid = uuid
	dsk.Metadata = info

	err = dsk.CommitFields(db, set.NewSet("uuid", "metadata"))
	if err != nil {
		return
	}

	return
}

func writeFsQcow(db *database.Database, dsk *disk.Disk) (err error) {
	diskPath, err := dsk.GetPath(db)
	if err != nil {
		return
	}

	err = utils.ExistsMkdir(filepath.Dir(diskPath), 0755)
	if err != nil {
		return
	}

	err = utils.Exec("", "qemu-img", "create",
		"-f", "qcow2", diskPath, fmt.Sprintf("%dG", dsk.Size))
	if err != nil {
		return
	}

	err = utils.Chmod(diskPath, 0600)
	if err != nil {
		return
	}

	logrus.WithFields(logrus.Fields{
		"disk_id":     dsk.Id.Hex(),
		"disk_path":   diskPath,
		"file_system": dsk.FileSystem,
	}).Info("data: Formatting qcow disk")

	sess, err := guestfs.Open(diskPath, guestfs.Qcow2, false)
	if err != nil {
		return
	}

	uuid, info, err := formatDevice(sess, dsk)

	closeErr := sess.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		return
	}

	err = commitFormat(db, dsk, uuid, info)
	if err != nil {
		return
	}

	return
}

func writeFsLvm(db *database.Database, dsk *disk.Disk,
	pl *pool.Pool) (err error) {

	vgName := pl.VgName
	lvName := dsk.Id.Hex()

	if _, _, e := parseFileSystem(dsk.FileSystem); e != nil {
		err = e
		return
	}

	lckr := lock.NewLvmLocker(vgName, lvName)
	acquired, err := lckr.Lock(db)
	if err != nil {
		return
	}

	if !acquired {
		err = &lock.LvmLockedError{
			errors.New("data: Failed to acquire LVM lock"),
		}
		return
	}
	defer func() {
		err2 := lckr.Unlock(db)
		if err2 != nil {
			logrus.WithFields(logrus.Fields{
				"error": err2,
			}).Error("data: Failed to unlock lvm")
		}
	}()

	err = lvm.CreateLv(vgName, lvName, dsk.Size)
	if err != nil {
		return
	}

	err = lvm.ActivateLv(vgName, lvName)
	if err != nil {
		return
	}

	defer func() {
		err2 := lvm.DeactivateLv(vgName, lvName)
		if err2 != nil {
			logrus.WithFields(logrus.Fields{
				"vg_name": vgName,
				"lv_name": lvName,
				"error":   err2,
			}).Error("data: Failed to deactivate lvm disk")
			if err == nil {
				err = err2
			}
		}
	}()

	diskPath := poolLvPath(vgName, lvName)

	logrus.WithFields(logrus.Fields{
		"disk_id":     dsk.Id.Hex(),
		"disk_path":   diskPath,
		"file_system": dsk.FileSystem,
	}).Info("data: Formatting lvm disk")

	sess, err := guestfs.Open(diskPath, guestfs.Raw, false)
	if err != nil {
		return
	}

	uuid, info, err := formatDevice(sess, dsk)

	closeErr := sess.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		return
	}

	err = commitFormat(db, dsk, uuid, info)
	if err != nil {
		return
	}

	return
}

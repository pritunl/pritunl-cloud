package data

import (
	"fmt"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/guestfs"
	"github.com/pritunl/pritunl-cloud/lock"
	"github.com/pritunl/pritunl-cloud/lvm"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/pool"
	"github.com/pritunl/pritunl-cloud/qcow"
	"github.com/sirupsen/logrus"
)

type partitionDevice struct {
	number int
	path   string
	size   int64
}

func getDiskSizeQcow(dsk *disk.Disk) (size int, err error) {
	return qcow.Size(paths.GetDiskPath(dsk.Id))
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
	err = qcow.Expand(dskPth, expandSize)
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
		err = &lock.LvmLockedError{
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

func checkFsExt(sess *guestfs.Session, devPath string) (err error) {
	err = sess.E2fsck(devPath)
	if err != nil {
		return
	}

	return
}

func growFsXfs(sess *guestfs.Session, devPath string) (err error) {
	err = sess.Mount(devPath, "/")
	if err != nil {
		return
	}

	err = sess.XfsGrowfs("/")
	if err != nil {
		return
	}

	err = sess.Umount("/")
	if err != nil {
		return
	}

	return
}

func growFs(sess *guestfs.Session, devPath, fs string) (err error) {
	switch fs {
	case disk.Xfs:
		err = growFsXfs(sess, devPath)
		break
	case disk.Ext4:
		err = checkFsExt(sess, devPath)
		if err != nil {
			return
		}

		err = sess.Resize2fs(devPath)
		break
	default:
		err = &errortypes.ParseError{
			errors.Newf("data: Cannot grow %s file system", fs),
		}
	}

	return
}

func shrinkFs(sess *guestfs.Session, devPath, fs string, size int64) (
	err error) {

	if fs != disk.Ext4 {
		err = &errortypes.ParseError{
			errors.Newf("data: Cannot shrink %s file system", fs),
		}
		return
	}

	err = checkFsExt(sess, devPath)
	if err != nil {
		return
	}

	err = sess.Resize2fsSize(devPath, size)
	if err != nil {
		return
	}

	return
}

func setPartitionSize(sess *guestfs.Session, devPath string,
	dev *partitionDevice, size int64, fill bool, sectorSize int) (
	err error) {

	line := ",+"
	if !fill {
		line = fmt.Sprintf(",%d", size/int64(sectorSize))
	}

	err = sess.SfdiskN(devPath, dev.number, line)
	if err != nil {
		return
	}

	partitions, err := sess.PartList(devPath)
	if err != nil {
		return
	}

	newSize := int64(-1)
	for _, partition := range partitions {
		if partition.Number == dev.number {
			newSize = partition.Size
			break
		}
	}

	if newSize < 0 {
		err = &errortypes.WriteError{
			errors.Newf("data: Partition %d missing after resize",
				dev.number),
		}
		return
	}

	if (fill && newSize < size) || (!fill && newSize != size) {
		err = &errortypes.WriteError{
			errors.Newf("data: Partition %d is %d bytes after resize, "+
				"expected %d", dev.number, newSize, size),
		}
		return
	}

	dev.size = newSize

	return
}

func resizeDevice(sess *guestfs.Session, dsk *disk.Disk) (err error) {
	meta := dsk.Metadata
	req := dsk.ActionData
	if meta == nil || req == nil {
		err = &errortypes.ParseError{
			errors.New("data: Disk resize request missing"),
		}
		return
	}

	devPath := sess.Device()

	sectorSize, err := sess.SectorSize(devPath)
	if err != nil {
		return
	}

	diskSize, err := sess.BlockSize(devPath)
	if err != nil {
		return
	}

	partDevs, err := sess.ListPartitions()
	if err != nil {
		return
	}

	devs := map[int]*partitionDevice{
		0: {
			path: devPath,
			size: diskSize,
		},
	}

	if len(partDevs) > 0 {
		partitions, e := sess.PartList(devPath)
		if e != nil {
			err = e
			return
		}

		for _, partition := range partitions {
			devs[partition.Number] = &partitionDevice{
				number: partition.Number,
				path:   partitionPath(devPath, partition.Number),
				size:   partition.Size,
			}
		}
	}

	parts := map[int]*disk.Partition{}
	vgs := map[string]*disk.VolumeGroup{}
	vgPvs := map[string]string{}
	for _, part := range meta.Partitions {
		dev := devs[part.Number]
		if dev == nil {
			err = &errortypes.WriteError{
				errors.Newf("data: Partition %d not found, "+
					"disk scan required", part.Number),
			}
			return
		}

		if dev.size != part.Size {
			err = &errortypes.WriteError{
				errors.Newf("data: Partition %d size changed, "+
					"disk scan required", part.Number),
			}
			return
		}

		parts[part.Number] = part

		if part.VolumeGroup != nil && part.VolumeGroup.Name != "" {
			vgs[part.VolumeGroup.Name] = part.VolumeGroup
			vgPvs[part.VolumeGroup.Name] = dev.path
		}
	}

	active := map[string]bool{}
	activate := func(vgName string) (err error) {
		if active[vgName] {
			return
		}

		err = sess.VgActivate(vgName, true)
		if err != nil {
			return
		}
		active[vgName] = true

		return
	}

	release := func() (err error) {
		err = sess.UmountAll()
		if err != nil {
			return
		}

		err = sess.VgActivateAll(false)
		if err != nil {
			return
		}
		active = map[string]bool{}

		return
	}

	findVolume := func(rv *disk.ResizeVolume) (
		vg *disk.VolumeGroup, lv *disk.LogicalVolume, err error) {

		vg = vgs[rv.VolumeGroup]
		if vg != nil {
			for _, volume := range vg.Volumes {
				if volume.Name == rv.Name {
					lv = volume
					break
				}
			}
		}

		if vg == nil || lv == nil {
			err = &errortypes.WriteError{
				errors.Newf("data: Volume %s/%s not found, "+
					"disk scan required", rv.VolumeGroup, rv.Name),
			}
			return
		}

		if !disk.IsManagedVgName(vg.Name) {
			err = &errortypes.WriteError{
				errors.Newf("data: Volume group %s cannot be resized",
					vg.Name),
			}
			return
		}

		return
	}

	for _, rv := range req.Volumes {
		if rv.Fill {
			continue
		}

		vg, lv, e := findVolume(rv)
		if e != nil {
			err = e
			return
		}

		if rv.Size >= lv.Size {
			continue
		}

		logrus.WithFields(logrus.Fields{
			"disk_id":  dsk.Id.Hex(),
			"volume":   vg.Name + "/" + lv.Name,
			"size":     lv.Size,
			"new_size": rv.Size,
		}).Info("data: Shrinking logical volume")

		err = activate(vg.Name)
		if err != nil {
			return
		}

		lvPath := lvDevicePath(vg.Name, lv.Name)

		err = shrinkFs(sess, lvPath, lv.FileSystem, rv.Size)
		if err != nil {
			return
		}

		err = sess.LvResize(lvPath, rv.Size/mib)
		if err != nil {
			return
		}

		err = growFs(sess, lvPath, lv.FileSystem)
		if err != nil {
			return
		}
	}

	err = release()
	if err != nil {
		return
	}

	for pass := 0; pass < 2; pass++ {
		for _, rp := range req.Partitions {
			part := parts[rp.Number]
			dev := devs[rp.Number]
			if part == nil || dev == nil || rp.Number <= 0 {
				err = &errortypes.WriteError{
					errors.Newf("data: Partition %d not found, "+
						"disk scan required", rp.Number),
				}
				return
			}

			shrink := !rp.Fill && rp.Size < part.Size
			if (pass == 0) != shrink {
				continue
			}
			if !shrink && !rp.Fill && rp.Size <= part.Size {
				continue
			}

			logrus.WithFields(logrus.Fields{
				"disk_id":   dsk.Id.Hex(),
				"partition": rp.Number,
				"size":      part.Size,
				"new_size":  rp.Size,
				"fill":      rp.Fill,
			}).Info("data: Resizing partition")

			if shrink {
				if part.FileSystem == disk.Lvm {
					err = sess.PvResizeSize(dev.path, rp.Size)
				} else {
					err = shrinkFs(sess, dev.path, part.FileSystem, rp.Size)
				}
				if err != nil {
					return
				}

				err = release()
				if err != nil {
					return
				}

				err = setPartitionSize(sess, devPath, dev, rp.Size, false,
					sectorSize)
				if err != nil {
					return
				}
			} else {
				err = release()
				if err != nil {
					return
				}

				err = setPartitionSize(sess, devPath, dev, rp.Size, rp.Fill,
					sectorSize)
				if err != nil {
					return
				}

				if part.FileSystem == disk.Lvm {
					err = sess.PvResize(dev.path)
				} else {
					err = growFs(sess, dev.path, part.FileSystem)
				}
				if err != nil {
					return
				}
			}
		}
	}

	for _, rv := range req.Volumes {
		vg, lv, e := findVolume(rv)
		if e != nil {
			err = e
			return
		}

		if !rv.Fill && rv.Size <= lv.Size {
			continue
		}

		logrus.WithFields(logrus.Fields{
			"disk_id":  dsk.Id.Hex(),
			"volume":   vg.Name + "/" + lv.Name,
			"size":     lv.Size,
			"new_size": rv.Size,
			"fill":     rv.Fill,
		}).Info("data: Growing logical volume")

		err = activate(vg.Name)
		if err != nil {
			return
		}

		lvPath := lvDevicePath(vg.Name, lv.Name)

		if rv.Fill {
			err = sess.LvResizeFree(lvPath, 100)
		} else {
			err = sess.LvResize(lvPath, rv.Size/mib)
		}
		if err != nil {
			return
		}

		err = growFs(sess, lvPath, lv.FileSystem)
		if err != nil {
			return
		}
	}

	err = release()
	if err != nil {
		return
	}

	err = sess.Sync()
	if err != nil {
		return
	}

	return
}

func resizePath(dsk *disk.Disk, pth, format string) (err error) {
	sess, err := guestfs.Open(pth, format, false)
	if err != nil {
		return
	}

	var info *disk.Metadata

	err = resizeDevice(sess, dsk)
	if err == nil {
		info, err = scanDevice(sess)
	}

	closeErr := sess.Close()
	if err == nil {
		err = closeErr
	}

	if err == nil {
		dsk.Metadata = info
		return
	}

	info, e := scanPath(pth, format)
	if e != nil {
		logrus.WithFields(logrus.Fields{
			"disk_id": dsk.Id.Hex(),
			"error":   e,
		}).Error("data: Failed to scan disk after resize")
		dsk.Metadata = nil
	} else {
		dsk.Metadata = info
	}

	return
}

func resizeDiskQcow(dsk *disk.Disk) (err error) {
	diskPath := paths.GetDiskPath(dsk.Id)

	err = resizePath(dsk, diskPath, guestfs.Qcow2)
	if err != nil {
		return
	}

	return
}

func resizeDiskLvm(db *database.Database, dsk *disk.Disk) (err error) {
	pl, err := pool.Get(db, dsk.Pool)
	if err != nil {
		return
	}

	vgName := pl.VgName
	lvName := dsk.Id.Hex()

	acquired, err := lock.LvmLock(db, vgName, lvName)
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
		err2 := lvm.DeactivateLv(vgName, lvName)
		if err2 != nil {
			logrus.WithFields(logrus.Fields{
				"vg_name": vgName,
				"lv_name": lvName,
				"error":   err2,
			}).Error("data: Failed to deactivate lvm disk")
		}
	}()

	err = resizePath(dsk, poolLvPath(vgName, lvName), guestfs.Raw)
	if err != nil {
		return
	}

	return
}

func ResizeDisk(db *database.Database, dsk *disk.Disk) (err error) {
	switch dsk.Type {
	case disk.Lvm:
		err = resizeDiskLvm(db, dsk)
		if err != nil {
			return
		}
		break
	case "", disk.Qcow2:
		err = resizeDiskQcow(dsk)
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

package data

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/lock"
	"github.com/pritunl/pritunl-cloud/lvm"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/pool"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/sirupsen/logrus"
)

type diskInfo struct {
	Filename    string `json:"filename"`
	Format      string `json:"format"`
	ActualSize  int    `json:"actual-size"`
	VirtualSize int    `json:"virtual-size"`
}

type partitionDevice struct {
	number int
	kname  string
	path   string
	size   int64
}

func getQcowSize(pth string) (size int, err error) {
	output, err := utils.ExecOutput("",
		"qemu-img", "info", "--output=json", pth)
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

func getDiskSizeQcow(dsk *disk.Disk) (size int, err error) {
	return getQcowSize(paths.GetDiskPath(dsk.Id))
}

func expandDiskQcow(db *database.Database, dsk *disk.Disk) (err error) {
	dskPth := paths.GetDiskPath(dsk.Id)

	logrus.WithFields(logrus.Fields{
		"disk_id":   dsk.Id.Hex(),
		"disk_path": dskPth,
		"new_size":  dsk.NewSize,
	}).Info("data: Expanding qcow disk")

	nbdLock.Lock()
	defer nbdLock.Unlock()

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

func getPartitionDevices(devPath string) (
	devs map[int]*partitionDevice, err error) {

	dev, err := getBlockDevice(devPath)
	if err != nil {
		return
	}

	devs = map[int]*partitionDevice{
		0: {
			kname: dev.KName,
			path:  devPath,
			size:  int64(dev.Size),
		},
	}

	for _, child := range dev.Children {
		if child.Type != "part" {
			continue
		}

		number, _, e := getPartitionOffset(child.KName)
		if e != nil {
			err = e
			return
		}

		devs[number] = &partitionDevice{
			number: number,
			kname:  child.KName,
			path:   filepath.Join("/dev", child.KName),
			size:   int64(child.Size),
		}
	}

	return
}

func getPartitionSize(kname string) (size int64, err error) {
	data, err := os.ReadFile(filepath.Join("/sys/class/block", kname, "size"))
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrapf(err, "data: Failed to read size of %s", kname),
		}
		return
	}

	sectors, err := strconv.ParseInt(
		strings.TrimSpace(string(data)), 10, 64)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrapf(err, "data: Failed to parse size of %s", kname),
		}
		return
	}

	size = sectors * 512

	return
}

func getSectorSize(kname string) int {
	data, err := os.ReadFile(filepath.Join(
		"/sys/class/block", kname, "queue", "logical_block_size"))
	if err != nil {
		return 512
	}

	size, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || size <= 0 {
		return 512
	}

	return size
}

func getBlockHolders(kname string) (holders []string, err error) {
	sysPath := filepath.Join("/sys/class/block", kname)

	entries, err := os.ReadDir(filepath.Join(sysPath, "holders"))
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrapf(err, "data: Failed to read holders of %s", kname),
		}
		return
	}

	for _, entry := range entries {
		holder := entry.Name()

		name, e := os.ReadFile(filepath.Join(
			"/sys/class/block", holder, "dm", "name"))
		if e == nil && strings.TrimSpace(string(name)) != "" {
			holder = fmt.Sprintf("%s (%s)", holder,
				strings.TrimSpace(string(name)))
		}

		holders = append(holders, "held by "+holder)
	}

	devData, err := os.ReadFile(filepath.Join(sysPath, "dev"))
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrapf(err,
				"data: Failed to read device number of %s", kname),
		}
		return
	}
	devNum := strings.TrimSpace(string(devData))

	mountData, err := os.ReadFile("/proc/self/mountinfo")
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "data: Failed to read mount table"),
		}
		return
	}

	for _, line := range strings.Split(string(mountData), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 5 || fields[2] != devNum {
			continue
		}

		holders = append(holders, "mounted on "+fields[4])
	}

	return
}

func execStatus(name string, arg ...string) (status int, err error) {
	cmd := exec.Command(name, arg...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if !ok {
			err = &errortypes.ExecError{
				errors.Wrapf(err, "data: Failed to exec '%s'", name),
			}
			return
		}

		status = exitErr.ExitCode()
		err = nil

		logrus.WithFields(logrus.Fields{
			"cmd":    name,
			"arg":    arg,
			"status": status,
			"output": string(output),
		}).Info("data: Command exited with status")
	}

	return
}

func checkFsExt(devPath string) (err error) {
	status, err := execStatus("e2fsck", "-f", "-p", devPath)
	if err != nil {
		return
	}

	if status > 2 {
		err = &errortypes.WriteError{
			errors.Newf("data: File system check failed with status %d",
				status),
		}
		return
	}

	return
}

func resizeFsExt(devPath string, size int64) (err error) {
	err = checkFsExt(devPath)
	if err != nil {
		return
	}

	args := []string{devPath}
	if size > 0 {
		args = append(args, fmt.Sprintf("%dK", size/1024))
	}

	_, err = utils.ExecCombinedOutputLogged(nil, "resize2fs", args...)
	if err != nil {
		return
	}

	return
}

func growFsXfs(devPath string) (err error) {
	err = utils.ExistsMkdir(paths.GetTempPath(), 0755)
	if err != nil {
		return
	}

	mountPath := filepath.Join(paths.GetTempPath(),
		fmt.Sprintf("resize-%s", bson.NewObjectID().Hex()))

	err = utils.ExistsMkdir(mountPath, 0700)
	if err != nil {
		return
	}
	defer os.Remove(mountPath)

	_, err = utils.ExecCombinedOutputLogged(nil, "mount", "-t", "xfs",
		"-o", "nouuid,noexec,nosuid,nodev", devPath, mountPath)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(nil, "xfs_growfs", mountPath)

	_, e := utils.ExecCombinedOutputLogged(nil, "umount", mountPath)
	if e != nil && err == nil {
		err = e
	}

	return
}

func growFs(devPath, fs string) (err error) {
	switch fs {
	case disk.Xfs:
		err = growFsXfs(devPath)
		break
	case disk.Ext4:
		err = resizeFsExt(devPath, 0)
		break
	default:
		err = &errortypes.ParseError{
			errors.Newf("data: Cannot grow %s file system", fs),
		}
	}

	return
}

func shrinkFs(devPath, fs string, size int64) (err error) {
	if fs != disk.Ext4 {
		err = &errortypes.ParseError{
			errors.Newf("data: Cannot shrink %s file system", fs),
		}
		return
	}

	err = resizeFsExt(devPath, size)
	if err != nil {
		return
	}

	return
}

func setPartitionSize(devPath string, dev *partitionDevice, size int64,
	fill bool, sectorSize int) (err error) {

	input := ",+\n"
	if !fill {
		input = fmt.Sprintf(",%d\n", size/int64(sectorSize))
	}

	utils.Exec("", "udevadm", "settle", "--timeout=5")

	_, err = utils.ExecInputOutputCombindLogged(input, "sfdisk",
		"-N", strconv.Itoa(dev.number), devPath)
	if err != nil {
		return
	}

	utils.ExecCombinedOutputLogged(nil, "partprobe", devPath)
	utils.Exec("", "udevadm", "settle", "--timeout=5")

	newSize, err := getPartitionSize(dev.kname)
	if err != nil {
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

func resizePv(pvPath string, size int64) (err error) {
	if size == 0 {
		_, err = utils.ExecCombinedOutputLogged(nil, "pvresize",
			"--devices", pvPath, pvPath)
	} else {
		_, err = utils.ExecCombinedOutputLogged(nil, "pvresize",
			"--devices", pvPath, "-y",
			"--setphysicalvolumesize", fmt.Sprintf("%db", size), pvPath)
	}
	if err != nil {
		return
	}

	return
}

func setVgActive(pvPath, vgName string, active bool) (err error) {
	flag := "-an"
	if active {
		flag = "-ay"
	}

	_, err = utils.ExecCombinedOutputLogged(nil, "vgchange",
		"--devices", pvPath, flag, vgName)
	if err != nil {
		return
	}

	utils.Exec("", "udevadm", "settle", "--timeout=5")

	return
}

func lvDevicePath(vgName, lvName string) string {
	return fmt.Sprintf("/dev/%s/%s", vgName, lvName)
}

func extendLv(pvPath, vgName, lvName string, size int64,
	fill bool) (err error) {

	args := []string{"--devices", pvPath}
	if fill {
		args = append(args, "-l", "+100%FREE")
	} else {
		args = append(args, "-L", fmt.Sprintf("%db", size))
	}
	args = append(args, fmt.Sprintf("%s/%s", vgName, lvName))

	_, err = utils.ExecCombinedOutputLogged(nil, "lvextend", args...)
	if err != nil {
		return
	}

	return
}

func reduceLv(pvPath, vgName, lvName string, size int64) (err error) {
	_, err = utils.ExecCombinedOutputLogged(nil, "lvreduce",
		"--devices", pvPath, "-y", "-L", fmt.Sprintf("%db", size),
		fmt.Sprintf("%s/%s", vgName, lvName))
	if err != nil {
		return
	}

	return
}

func resizeDevice(dsk *disk.Disk, devPath string) (err error) {
	meta := dsk.Metadata
	req := dsk.ActionData
	if meta == nil || req == nil {
		err = &errortypes.ParseError{
			errors.New("data: Disk resize request missing"),
		}
		return
	}

	devs, err := getPartitionDevices(devPath)
	if err != nil {
		return
	}

	sectorSize := getSectorSize(devs[0].kname)

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

		if part.VolumeGroup != nil {
			vgs[part.VolumeGroup.Name] = part.VolumeGroup
			vgPvs[part.VolumeGroup.Name] = dev.path
		}
	}

	active := map[string]bool{}
	activate := func(vgName string) (err error) {
		if active[vgName] {
			return
		}

		err = setVgActive(vgPvs[vgName], vgName, true)
		if err != nil {
			return
		}
		active[vgName] = true

		return
	}

	deactivate := func() (err error) {
		for vgName, pvPath := range vgPvs {
			e := setVgActive(pvPath, vgName, false)
			if e != nil && err == nil {
				err = e
			}
		}
		active = map[string]bool{}

		return
	}
	defer func() {
		e := deactivate()
		if e != nil {
			logrus.WithFields(logrus.Fields{
				"disk_id": dsk.Id.Hex(),
				"error":   e,
			}).Error("data: Failed to deactivate volume group")
		}
	}()

	release := func() (err error) {
		inUse := []string{}
		for attempt := 0; attempt < 5; attempt++ {
			if attempt > 0 {
				time.Sleep(1 * time.Second)
			}

			utils.Exec("", "udevadm", "settle", "--timeout=5")

			e := deactivate()
			if e != nil {
				logrus.WithFields(logrus.Fields{
					"disk_id": dsk.Id.Hex(),
					"error":   e,
				}).Warn("data: Failed to deactivate volume group")
			}

			inUse = []string{}
			for _, dev := range devs {
				if dev.number == 0 {
					continue
				}

				holders, e := getBlockHolders(dev.kname)
				if e != nil {
					err = e
					return
				}

				for _, holder := range holders {
					inUse = append(inUse, fmt.Sprintf(
						"partition %d %s", dev.number, holder))
				}
			}

			if len(inUse) == 0 {
				return
			}
		}

		sort.Strings(inUse)
		err = &errortypes.WriteError{
			errors.Newf("data: Disk in use, %s", strings.Join(inUse, ", ")),
		}
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

		err = shrinkFs(lvPath, lv.FileSystem, rv.Size)
		if err != nil {
			return
		}

		err = reduceLv(vgPvs[vg.Name], vg.Name, lv.Name, rv.Size)
		if err != nil {
			return
		}

		err = resizeFsExt(lvPath, 0)
		if err != nil {
			return
		}
	}

	err = deactivate()
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

			err = release()
			if err != nil {
				return
			}

			if shrink {
				if part.FileSystem == disk.Lvm {
					err = resizePv(dev.path, rp.Size)
				} else {
					err = shrinkFs(dev.path, part.FileSystem, rp.Size)
				}
				if err != nil {
					return
				}

				err = setPartitionSize(devPath, dev, rp.Size, false,
					sectorSize)
				if err != nil {
					return
				}
			} else {
				err = setPartitionSize(devPath, dev, rp.Size, rp.Fill,
					sectorSize)
				if err != nil {
					return
				}

				if part.FileSystem == disk.Lvm {
					err = resizePv(dev.path, 0)
				} else {
					err = growFs(dev.path, part.FileSystem)
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

		err = extendLv(vgPvs[vg.Name], vg.Name, lv.Name, rv.Size, rv.Fill)
		if err != nil {
			return
		}

		err = growFs(lvDevicePath(vg.Name, lv.Name), lv.FileSystem)
		if err != nil {
			return
		}
	}

	err = deactivate()
	if err != nil {
		return
	}

	return
}

func rescanDevice(dsk *disk.Disk, devPath string) {
	info, err := scanFsDevice(devPath)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"disk_id": dsk.Id.Hex(),
			"error":   err,
		}).Error("data: Failed to scan disk after resize")
		return
	}

	dsk.Metadata = info
}

func resizeDiskQcow(dsk *disk.Disk) (err error) {
	ndbPath := settings.Hypervisor.NbdPath
	diskPath := paths.GetDiskPath(dsk.Id)

	nbdLock.Lock()
	defer func() {
		utils.Exec("", "sync")
		utils.Exec("", "qemu-nbd", "--disconnect", ndbPath)
		nbdLock.Unlock()
	}()

	_, err = utils.ExecCombinedOutputLogged(
		nil, "modprobe", "nbd")
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil, "qemu-nbd", "--disconnect", ndbPath)
	if err != nil {
		return
	}

	time.Sleep(300 * time.Millisecond)

	_, err = utils.ExecCombinedOutputLogged(
		nil, "qemu-nbd", "--connect", ndbPath, diskPath)
	if err != nil {
		return
	}

	time.Sleep(1 * time.Second)

	err = resizeDevice(dsk, ndbPath)
	rescanDevice(dsk, ndbPath)

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
		err2 := lvm.DeactivateLv(vgName, lvName)
		if err2 != nil {
			logrus.WithFields(logrus.Fields{
				"vg_name": vgName,
				"lv_name": lvName,
				"error":   err2,
			}).Error("data: Failed to deactivate lvm disk")
		}
	}()

	diskPath := filepath.Join("/dev/mapper",
		fmt.Sprintf("%s-%s", vgName, lvName))

	err = resizeDevice(dsk, diskPath)
	rescanDevice(dsk, diskPath)

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

package data

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/lock"
	"github.com/pritunl/pritunl-cloud/lvm"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/pool"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/systemd"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/sirupsen/logrus"
)

type lsblkNumber int64

func (n *lsblkNumber) UnmarshalJSON(data []byte) (err error) {
	val := strings.Trim(strings.TrimSpace(string(data)), "\"")
	if val == "" || val == "null" {
		*n = 0
		return
	}

	num, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "data: Failed to parse lsblk number"),
		}
		return
	}

	*n = lsblkNumber(num)

	return
}

type lsblkDevice struct {
	Name     string         `json:"name"`
	KName    string         `json:"kname"`
	Type     string         `json:"type"`
	Size     lsblkNumber    `json:"size"`
	FsType   string         `json:"fstype"`
	Uuid     string         `json:"uuid"`
	Label    string         `json:"label"`
	PtType   string         `json:"pttype"`
	Children []*lsblkDevice `json:"children"`
}

type lsblkReport struct {
	BlockDevices []*lsblkDevice `json:"blockdevices"`
}

type lvmReport struct {
	Report []*lvmReportGroup `json:"report"`
}

type lvmReportGroup struct {
	Pv []map[string]interface{} `json:"pv"`
	Lv []map[string]interface{} `json:"lv"`
}

func lvmReportString(row map[string]interface{}, key string) string {
	switch val := row[key].(type) {
	case string:
		return strings.TrimSpace(val)
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	default:
		return ""
	}
}

func lvmReportInt(row map[string]interface{}, key string) int64 {
	val := lvmReportString(row, key)
	if val == "" {
		return 0
	}

	num, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0
	}

	return int64(num)
}

func parseLvmReport(output string) (reprt *lvmReport, err error) {
	reprt = &lvmReport{}

	err = json.Unmarshal([]byte(output), reprt)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "data: Failed to parse lvm report"),
		}
		return
	}

	return
}

func fsType(fs string) string {
	if fs == lvmMember {
		return disk.Lvm
	}
	return fs
}

func getBlockDevice(devPath string) (dev *lsblkDevice, err error) {
	output, err := utils.ExecOutputLogged(nil, "lsblk",
		"--json", "--bytes",
		"--output", "NAME,KNAME,TYPE,SIZE,FSTYPE,UUID,LABEL,PTTYPE",
		devPath)
	if err != nil {
		return
	}

	reprt := &lsblkReport{}
	err = json.Unmarshal([]byte(output), reprt)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "data: Failed to parse lsblk report"),
		}
		return
	}

	if len(reprt.BlockDevices) == 0 {
		err = &errortypes.ReadError{
			errors.Newf("data: Block device %s not found", devPath),
		}
		return
	}

	dev = reprt.BlockDevices[0]

	return
}

func getPartitionOffset(kname string) (number int, start int64, err error) {
	sysPath := filepath.Join("/sys/class/block", kname)

	numberData, err := os.ReadFile(filepath.Join(sysPath, "partition"))
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrapf(err,
				"data: Failed to read partition number for %s", kname),
		}
		return
	}

	number, err = strconv.Atoi(strings.TrimSpace(string(numberData)))
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrapf(err,
				"data: Failed to parse partition number for %s", kname),
		}
		return
	}

	startData, err := os.ReadFile(filepath.Join(sysPath, "start"))
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrapf(err,
				"data: Failed to read partition start for %s", kname),
		}
		return
	}

	startSectors, err := strconv.ParseInt(
		strings.TrimSpace(string(startData)), 10, 64)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrapf(err,
				"data: Failed to parse partition start for %s", kname),
		}
		return
	}

	start = startSectors * 512

	return
}

func scanVolumeGroup(pvPath string) (vg *disk.VolumeGroup, err error) {
	output, err := utils.ExecOutputLogged(nil, "pvs",
		"--devices", pvPath,
		"--reportformat", "json",
		"--units", "b", "--nosuffix",
		"--options", "pv_name,vg_name,vg_size,vg_free,vg_extent_size",
		pvPath)
	if err != nil {
		return
	}

	reprt, err := parseLvmReport(output)
	if err != nil {
		return
	}

	var pvRow map[string]interface{}
	for _, group := range reprt.Report {
		if len(group.Pv) > 0 {
			pvRow = group.Pv[0]
			break
		}
	}

	if pvRow == nil {
		err = &errortypes.ReadError{
			errors.Newf("data: Physical volume %s not found", pvPath),
		}
		return
	}

	vgName := lvmReportString(pvRow, "vg_name")

	vg = &disk.VolumeGroup{
		Name:       vgName,
		Size:       lvmReportInt(pvRow, "vg_size"),
		Free:       lvmReportInt(pvRow, "vg_free"),
		ExtentSize: lvmReportInt(pvRow, "vg_extent_size"),
		Volumes:    []*disk.LogicalVolume{},
	}

	if vgName == "" {
		return
	}

	output, err = utils.ExecOutputLogged(nil, "lvs",
		"--devices", pvPath,
		"--reportformat", "json",
		"--units", "b", "--nosuffix",
		"--options", "lv_name,lv_size,lv_dm_path",
		vgName)
	if err != nil {
		return
	}

	reprt, err = parseLvmReport(output)
	if err != nil {
		return
	}

	activated := false
	if IsManagedVgName(vgName) {
		_, e := utils.ExecCombinedOutputLogged(nil, "vgchange",
			"--devices", pvPath, "-ay", vgName)
		if e != nil {
			logrus.WithFields(logrus.Fields{
				"vg_name": vgName,
				"error":   e,
			}).Warn("data: Failed to activate volume group for scan")
		} else {
			activated = true

			defer func() {
				_, e := utils.ExecCombinedOutputLogged(nil, "vgchange",
					"--devices", pvPath, "-an", vgName)
				if e != nil {
					logrus.WithFields(logrus.Fields{
						"vg_name": vgName,
						"error":   e,
					}).Error("data: Failed to deactivate volume group")
				}
			}()

			time.Sleep(100 * time.Millisecond)
		}
	}

	for _, group := range reprt.Report {
		for _, row := range group.Lv {
			lv := &disk.LogicalVolume{
				Name: lvmReportString(row, "lv_name"),
				Size: lvmReportInt(row, "lv_size"),
			}

			dmPath := lvmReportString(row, "lv_dm_path")
			if activated && dmPath != "" {
				dev, e := getBlockDevice(dmPath)
				if e != nil {
					logrus.WithFields(logrus.Fields{
						"vg_name": vgName,
						"lv_name": lv.Name,
						"error":   e,
					}).Warn("data: Failed to read logical volume")
				} else {
					lv.FileSystem = fsType(dev.FsType)
					lv.Uuid = dev.Uuid
					lv.Label = dev.Label
				}
			}

			vg.Volumes = append(vg.Volumes, lv)
		}
	}

	return
}

func scanFsDevice(devPath string) (info *disk.Metadata, err error) {
	utils.Exec("", "udevadm", "settle", "--timeout=5")

	dev, err := getBlockDevice(devPath)
	if err != nil {
		return
	}

	info = &disk.Metadata{
		Timestamp:  time.Now(),
		Size:       int64(dev.Size),
		Table:      dev.PtType,
		Partitions: []*disk.Partition{},
	}

	for _, child := range dev.Children {
		if child.Type != "part" {
			continue
		}

		number, start, e := getPartitionOffset(child.KName)
		if e != nil {
			err = e
			return
		}

		part := &disk.Partition{
			Number:     number,
			Start:      start,
			Size:       int64(child.Size),
			FileSystem: fsType(child.FsType),
			Uuid:       child.Uuid,
			Label:      utils.FilterStr(child.Label, 128),
		}

		if part.FileSystem == disk.Lvm {
			vg, e := scanVolumeGroup(filepath.Join("/dev", child.KName))
			if e != nil {
				logrus.WithFields(logrus.Fields{
					"device":    devPath,
					"partition": number,
					"error":     e,
				}).Warn("data: Failed to scan volume group")
			}
			part.VolumeGroup = vg
		}

		info.Partitions = append(info.Partitions, part)
	}

	if len(info.Partitions) == 0 && dev.FsType != "" {
		part := &disk.Partition{
			Size:       int64(dev.Size),
			FileSystem: fsType(dev.FsType),
			Uuid:       dev.Uuid,
			Label:      utils.FilterStr(dev.Label, 128),
		}

		if part.FileSystem == disk.Lvm {
			vg, e := scanVolumeGroup(devPath)
			if e != nil {
				logrus.WithFields(logrus.Fields{
					"device": devPath,
					"error":  e,
				}).Warn("data: Failed to scan volume group")
			}
			part.VolumeGroup = vg
		}

		info.Partitions = append(info.Partitions, part)
	}

	sort.Slice(info.Partitions, func(i, j int) bool {
		return info.Partitions[i].Start < info.Partitions[j].Start
	})

	return
}

func scanDiskQcow(dsk *disk.Disk) (info *disk.Metadata, err error) {
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

	info, err = scanFsDevice(ndbPath)
	if err != nil {
		return
	}

	return
}

func scanDiskLvm(db *database.Database, dsk *disk.Disk) (
	info *disk.Metadata, err error) {

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

	info, err = scanFsDevice(diskPath)
	if err != nil {
		return
	}

	return
}

func DiskInUse(db *database.Database, dsk *disk.Disk) (
	inUse bool, err error) {

	if dsk.Instance.IsZero() {
		return
	}

	inst, err := instance.Get(db, dsk.Instance)
	if err != nil {
		if _, ok := err.(*database.NotFoundError); ok {
			err = nil
			inUse = true
		}
		return
	}

	if inst.Action != instance.Stop {
		inUse = true
		return
	}

	state, _, err := systemd.GetState(paths.GetUnitName(dsk.Instance))
	if err != nil {
		return
	}

	switch state {
	case "inactive", "failed", "unknown", "":
		break
	default:
		inUse = true
		break
	}

	return
}

func ScanDisk(db *database.Database, dsk *disk.Disk) (err error) {
	var info *disk.Metadata

	switch dsk.Type {
	case disk.Lvm:
		info, err = scanDiskLvm(db, dsk)
		if err != nil {
			return
		}
		break
	case "", disk.Qcow2:
		info, err = scanDiskQcow(dsk)
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

	dsk.Metadata = info

	return
}

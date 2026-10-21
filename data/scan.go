package data

import (
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/guestfs"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/lock"
	"github.com/pritunl/pritunl-cloud/lvm"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/pool"
	"github.com/pritunl/pritunl-cloud/systemd"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/sirupsen/logrus"
)

const (
	labelLimit = 128
)

func fsType(fs string) string {
	switch fs {
	case guestfs.LvmMember:
		return disk.Lvm
	case "unknown":
		return ""
	}
	return fs
}

func partitionPath(dev string, number int) string {
	if len(dev) > 0 && dev[len(dev)-1] >= '0' && dev[len(dev)-1] <= '9' {
		return dev + "p" + strconv.Itoa(number)
	}
	return dev + strconv.Itoa(number)
}

func probeFileSystem(sess *guestfs.Session, device string) (
	fs, uuid, label string, err error) {

	output, ok, err := sess.RunOptional("vfs-type", device)
	if err != nil {
		return
	}
	if !ok {
		return
	}

	fs = fsType(strings.TrimSpace(output))
	if fs == "" {
		return
	}

	output, ok, err = sess.RunOptional("vfs-uuid", device)
	if err != nil {
		return
	}
	if ok {
		uuid = utils.FilterStr(strings.TrimSpace(output), labelLimit)
	}

	output, ok, err = sess.RunOptional("vfs-label", device)
	if err != nil {
		return
	}
	if ok {
		label = utils.FilterStr(strings.TrimRight(output, "\n"), labelLimit)
	}

	return
}

type volumeGroups struct {
	groups map[string]*disk.VolumeGroup
	byPv   map[string]string
}

func scanVolumeGroups(sess *guestfs.Session) (vgs *volumeGroups,
	err error) {

	vgs = &volumeGroups{
		groups: map[string]*disk.VolumeGroup{},
		byPv:   map[string]string{},
	}

	ok, err := sess.VgActivateAllOptional(true)
	if err != nil {
		return
	}
	if !ok {
		logrus.WithFields(logrus.Fields{
			"path": sess.Device(),
		}).Warn("data: Failed to activate volume groups for scan")
	}

	groups, err := sess.VgsFull()
	if err != nil {
		return
	}

	if len(groups) == 0 {
		return
	}

	volumes, err := sess.LvsFull()
	if err != nil {
		return
	}

	lvPaths, err := sess.Lvs()
	if err != nil {
		return
	}

	lvByName := map[string]string{}
	for _, lvPath := range lvPaths {
		fields := strings.Split(lvPath, "/")
		if len(fields) != 4 || fields[1] != "dev" {
			continue
		}
		lvByName[fields[2]+"/"+fields[3]] = lvPath
	}

	byUuid := map[string]*guestfs.LogicalVolume{}
	for i := range volumes {
		byUuid[volumes[i].Uuid] = &volumes[i]
	}

	for _, group := range groups {
		if group.Name == "" {
			continue
		}

		vg := &disk.VolumeGroup{
			Name:       group.Name,
			Size:       group.Size,
			Free:       group.Free,
			ExtentSize: group.ExtentSize,
			Volumes:    []*disk.LogicalVolume{},
		}

		pvUuids, e := sess.VgPvUuids(group.Name)
		if e != nil {
			err = e
			return
		}
		for _, pvUuid := range pvUuids {
			vgs.byPv[pvUuid] = group.Name
		}

		lvUuids, e := sess.VgLvUuids(group.Name)
		if e != nil {
			err = e
			return
		}

		for _, lvUuid := range lvUuids {
			volume := byUuid[lvUuid]
			if volume == nil {
				continue
			}

			lv := &disk.LogicalVolume{
				Name: volume.Name,
				Size: volume.Size,
			}

			lvPath := lvByName[group.Name+"/"+volume.Name]
			if lvPath != "" {
				fs, uuid, label, e := probeFileSystem(sess, lvPath)
				if e != nil {
					err = e
					return
				}

				lv.FileSystem = fs
				lv.Uuid = uuid
				lv.Label = label
			}

			vg.Volumes = append(vg.Volumes, lv)
		}

		sort.SliceStable(vg.Volumes, func(i, j int) bool {
			return vg.Volumes[i].Name < vg.Volumes[j].Name
		})

		vgs.groups[group.Name] = vg
	}

	return
}

func (v *volumeGroups) get(sess *guestfs.Session, device string) (
	vg *disk.VolumeGroup, err error) {

	vg = &disk.VolumeGroup{
		Volumes: []*disk.LogicalVolume{},
	}

	output, ok, err := sess.RunOptional("pvuuid", device)
	if err != nil {
		return
	}
	if !ok {
		return
	}

	vgName := v.byPv[guestfs.NormalizeUuid(output)]
	if vgName == "" {
		return
	}

	group := v.groups[vgName]
	if group != nil {
		vg = group
	}

	return
}

func scanDevice(sess *guestfs.Session) (info *disk.Metadata, err error) {
	dev := sess.Device()

	size, err := sess.BlockSize(dev)
	if err != nil {
		return
	}

	table, err := sess.PartTable(dev)
	if err != nil {
		return
	}

	partDevs, err := sess.ListPartitions()
	if err != nil {
		return
	}

	partitions := []guestfs.Partition{}
	if len(partDevs) > 0 {
		partitions, err = sess.PartList(dev)
		if err != nil {
			return
		}
	}

	vgs, err := scanVolumeGroups(sess)
	if err != nil {
		return
	}

	info = &disk.Metadata{
		Timestamp:  time.Now(),
		Size:       size,
		Table:      table,
		Partitions: []*disk.Partition{},
	}

	for _, partition := range partitions {
		partDev := partitionPath(dev, partition.Number)

		part := &disk.Partition{
			Number: partition.Number,
			Start:  partition.Start,
			Size:   partition.Size,
		}

		fs, uuid, label, e := probeFileSystem(sess, partDev)
		if e != nil {
			err = e
			return
		}

		part.FileSystem = fs
		part.Uuid = uuid
		part.Label = label

		if part.FileSystem == disk.Lvm {
			vg, e := vgs.get(sess, partDev)
			if e != nil {
				err = e
				return
			}
			part.VolumeGroup = vg
		}

		info.Partitions = append(info.Partitions, part)
	}

	if len(info.Partitions) == 0 {
		fs, uuid, label, e := probeFileSystem(sess, dev)
		if e != nil {
			err = e
			return
		}

		if fs != "" {
			part := &disk.Partition{
				Size:       size,
				FileSystem: fs,
				Uuid:       uuid,
				Label:      label,
			}

			if part.FileSystem == disk.Lvm {
				vg, e := vgs.get(sess, dev)
				if e != nil {
					err = e
					return
				}
				part.VolumeGroup = vg
			}

			info.Partitions = append(info.Partitions, part)
		}
	}

	sort.SliceStable(info.Partitions, func(i, j int) bool {
		return info.Partitions[i].Start < info.Partitions[j].Start
	})

	return
}

func scanPath(pth, format string) (info *disk.Metadata, err error) {
	sess, err := guestfs.Open(pth, format, true)
	if err != nil {
		return
	}
	defer sess.Close()

	info, err = scanDevice(sess)
	if err != nil {
		return
	}

	return
}

func scanDiskQcow(dsk *disk.Disk) (info *disk.Metadata, err error) {
	diskPath := paths.GetDiskPath(dsk.Id)

	exists, err := utils.ExistsFile(diskPath)
	if err != nil {
		return
	}

	if !exists {
		err = &errortypes.NotFoundError{
			errors.Newf("data: Disk image %s not found", diskPath),
		}
		return
	}

	info, err = scanPath(diskPath, guestfs.Qcow2)
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

	info, err = scanPath(poolLvPath(vgName, lvName), guestfs.Raw)
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

package guestfs

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

const (
	Gpt   = "gpt"
	Msdos = "msdos"

	LvmMember = "LVM2_member"
)

type Partition struct {
	Number int
	Start  int64
	End    int64
	Size   int64
}

type FileSystem struct {
	Device string
	Type   string
}

type VolumeGroup struct {
	Name       string
	Uuid       string
	Size       int64
	Free       int64
	ExtentSize int64
}

type LogicalVolume struct {
	Name string
	Uuid string
	Size int64
}

func NormalizeUuid(uuid string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(uuid),
		"-", ""))
}

func Quote(val string) string {
	buf := strings.Builder{}
	buf.WriteByte('"')
	for _, b := range []byte(val) {
		buf.WriteString(fmt.Sprintf("\\x%02x", b))
	}
	buf.WriteByte('"')
	return buf.String()
}

func (s *Session) ListDevices() (devices []string, err error) {
	output, err := s.Run("list-devices")
	if err != nil {
		return
	}

	devices = parseLines(output)

	return
}

func (s *Session) ListPartitions() (partitions []string, err error) {
	output, err := s.Run("list-partitions")
	if err != nil {
		return
	}

	partitions = parseLines(output)

	return
}

func (s *Session) ListFileSystems() (fileSystems []FileSystem, err error) {
	output, err := s.Run("list-filesystems")
	if err != nil {
		return
	}

	fileSystems = []FileSystem{}

	keys, pairs := parsePairs(output)
	for _, key := range keys {
		fileSystems = append(fileSystems, FileSystem{
			Device: key,
			Type:   pairs[key],
		})
	}

	return
}

func (s *Session) BlockSize(device string) (size int64, err error) {
	output, err := s.Run("blockdev-getsize64", device)
	if err != nil {
		return
	}

	size, err = parseInt(output)
	if err != nil {
		return
	}

	return
}

func (s *Session) SectorSize(device string) (size int, err error) {
	output, err := s.Run("blockdev-getss", device)
	if err != nil {
		return
	}

	num, err := parseInt(output)
	if err != nil {
		return
	}

	if num <= 0 {
		err = &errortypes.ParseError{
			errors.Newf("guestfs: Invalid sector size %d", num),
		}
		return
	}

	size = int(num)

	return
}

func (s *Session) PartTable(device string) (table string, err error) {
	output, ok, err := s.RunOptional("part-get-parttype", device)
	if err != nil {
		return
	}

	if !ok {
		return
	}

	table = parseValue(output)

	return
}

func (s *Session) PartList(device string) (
	partitions []Partition, err error) {

	output, err := s.Run("part-list", device)
	if err != nil {
		return
	}

	items, err := parseStructs(output)
	if err != nil {
		return
	}

	partitions = []Partition{}
	for _, item := range items {
		num, e := structInt(item, "part_num")
		if e != nil {
			err = e
			return
		}

		start, e := structInt(item, "part_start")
		if e != nil {
			err = e
			return
		}

		end, e := structInt(item, "part_end")
		if e != nil {
			err = e
			return
		}

		size, e := structInt(item, "part_size")
		if e != nil {
			err = e
			return
		}

		partitions = append(partitions, Partition{
			Number: int(num),
			Start:  start,
			End:    end,
			Size:   size,
		})
	}

	return
}

func (s *Session) PartInit(device, table string) (err error) {
	_, err = s.Run("part-init", device, table)
	if err != nil {
		return
	}

	return
}

func (s *Session) PartAdd(device string, startSect, endSect int64) (
	err error) {

	_, err = s.Run("part-add", device, "p",
		strconv.FormatInt(startSect, 10), strconv.FormatInt(endSect, 10))
	return
}

func (s *Session) SfdiskN(device string, number int, line string) (
	err error) {

	_, err = s.Run("sfdisk-N", device, strconv.Itoa(number), "0", "0", "0",
		Quote(line))
	if err != nil {
		return
	}

	return
}

func (s *Session) VfsType(device string) (typ string, err error) {
	output, err := s.Run("vfs-type", device)
	if err != nil {
		return
	}

	typ = parseValue(output)

	return
}

func (s *Session) VfsUuid(device string) (uuid string, err error) {
	output, err := s.Run("vfs-uuid", device)
	if err != nil {
		return
	}

	uuid = parseValue(output)

	return
}

func (s *Session) VfsLabel(device string) (label string, err error) {
	output, err := s.Run("vfs-label", device)
	if err != nil {
		return
	}

	label = strings.TrimRight(output, "\n")

	return
}

func (s *Session) Mkfs(fsType, device string) (err error) {
	_, err = s.Run("mkfs", fsType, device)
	return
}

func (s *Session) PvCreate(device string) (err error) {
	_, err = s.Run("pvcreate", device)
	return
}

func (s *Session) VgCreate(vgName, device string) (err error) {
	_, err = s.Run("vgcreate", vgName, device)
	return
}

func (s *Session) LvCreate(lvName, vgName string, sizeMb int64) (
	err error) {

	_, err = s.Run("lvcreate", lvName, vgName, strconv.FormatInt(sizeMb, 10))
	if err != nil {
		return
	}

	return
}

func (s *Session) LvCreateFree(lvName, vgName string, percent int) (
	err error) {

	_, err = s.Run("lvcreate-free", lvName, vgName, strconv.Itoa(percent))
	if err != nil {
		return
	}

	return
}

func (s *Session) Vgs() (names []string, err error) {
	output, err := s.Run("vgs")
	if err != nil {
		return
	}

	names = parseLines(output)

	return
}

func (s *Session) Lvs() (paths []string, err error) {
	output, err := s.Run("lvs")
	if err != nil {
		return
	}

	paths = parseLines(output)

	return
}

func (s *Session) Pvs() (devices []string, err error) {
	output, err := s.Run("pvs")
	if err != nil {
		return
	}

	devices = parseLines(output)

	return
}

func (s *Session) VgsFull() (groups []VolumeGroup, err error) {
	output, err := s.Run("vgs-full")
	if err != nil {
		return
	}

	items, err := parseStructs(output)
	if err != nil {
		return
	}

	groups = []VolumeGroup{}
	for _, item := range items {
		size, e := structInt(item, "vg_size")
		if e != nil {
			err = e
			return
		}

		free, e := structInt(item, "vg_free")
		if e != nil {
			err = e
			return
		}

		extentSize, e := structInt(item, "vg_extent_size")
		if e != nil {
			err = e
			return
		}

		groups = append(groups, VolumeGroup{
			Name:       item["vg_name"],
			Uuid:       NormalizeUuid(item["vg_uuid"]),
			Size:       size,
			Free:       free,
			ExtentSize: extentSize,
		})
	}

	return
}

func (s *Session) LvsFull() (volumes []LogicalVolume, err error) {
	output, err := s.Run("lvs-full")
	if err != nil {
		return
	}

	items, err := parseStructs(output)
	if err != nil {
		return
	}

	volumes = []LogicalVolume{}
	for _, item := range items {
		size, e := structInt(item, "lv_size")
		if e != nil {
			err = e
			return
		}

		volumes = append(volumes, LogicalVolume{
			Name: item["lv_name"],
			Uuid: NormalizeUuid(item["lv_uuid"]),
			Size: size,
		})
	}

	return
}

func (s *Session) PvUuid(device string) (uuid string, err error) {
	output, err := s.Run("pvuuid", device)
	if err != nil {
		return
	}

	uuid = NormalizeUuid(output)

	return
}

func (s *Session) VgPvUuids(vgName string) (uuids []string, err error) {
	output, err := s.Run("vgpvuuids", vgName)
	if err != nil {
		return
	}

	uuids = []string{}
	for _, line := range parseLines(output) {
		uuids = append(uuids, NormalizeUuid(line))
	}

	return
}

func (s *Session) VgLvUuids(vgName string) (uuids []string, err error) {
	output, err := s.Run("vglvuuids", vgName)
	if err != nil {
		return
	}

	uuids = []string{}
	for _, line := range parseLines(output) {
		uuids = append(uuids, NormalizeUuid(line))
	}

	return
}

func (s *Session) VgActivate(vgName string, active bool) (err error) {
	_, err = s.Run("vg-activate", strconv.FormatBool(active), vgName)
	if err != nil {
		return
	}

	return
}

func (s *Session) VgActivateAll(active bool) (err error) {
	_, err = s.Run("vg-activate-all", strconv.FormatBool(active))
	if err != nil {
		return
	}

	return
}

func (s *Session) VgActivateAllOptional(active bool) (ok bool, err error) {
	_, ok, err = s.RunOptional("vg-activate-all", strconv.FormatBool(active))
	if err != nil {
		return
	}

	return
}

func (s *Session) Mount(device, mountPoint string) (err error) {
	_, err = s.Run("mount", device, mountPoint)
	if err != nil {
		return
	}

	return
}

func (s *Session) Umount(mountPoint string) (err error) {
	_, err = s.Run("umount", mountPoint)
	if err != nil {
		return
	}

	return
}

func (s *Session) UmountAll() (err error) {
	_, err = s.Run("umount-all")
	if err != nil {
		return
	}

	return
}

func (s *Session) E2fsck(device string) (err error) {
	_, err = s.Run("e2fsck", device, "correct:true")
	if err != nil {
		return
	}

	return
}

func (s *Session) Resize2fs(device string) (err error) {
	_, err = s.Run("resize2fs", device)
	if err != nil {
		return
	}

	return
}

func (s *Session) Resize2fsSize(device string, size int64) (err error) {
	_, err = s.Run("resize2fs-size", device, strconv.FormatInt(size, 10))
	if err != nil {
		return
	}

	return
}

func (s *Session) XfsGrowfs(mountPoint string) (err error) {
	_, err = s.Run("xfs-growfs", mountPoint)
	if err != nil {
		return
	}

	return
}

func (s *Session) PvResize(device string) (err error) {
	_, err = s.Run("pvresize", device)
	if err != nil {
		return
	}

	return
}

func (s *Session) PvResizeSize(device string, size int64) (err error) {
	_, err = s.Run("pvresize-size", device, strconv.FormatInt(size, 10))
	if err != nil {
		return
	}

	return
}

func (s *Session) LvResize(lvPath string, sizeMb int64) (err error) {
	_, err = s.Run("lvresize", lvPath, strconv.FormatInt(sizeMb, 10))
	if err != nil {
		return
	}

	return
}

func (s *Session) LvResizeFree(lvPath string, percent int) (err error) {
	_, err = s.Run("lvresize-free", lvPath, strconv.Itoa(percent))
	if err != nil {
		return
	}

	return
}

func (s *Session) Sync() (err error) {
	_, err = s.Run("sync")
	if err != nil {
		return
	}

	return
}

package disk

import (
	"fmt"
	"sort"
	"strings"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

const (
	ResizeAlign   int64 = 1024 * 1024
	ResizeMinSize int64 = 1024 * 1024 * 1024
	gptReserve    int64 = 1024 * 1024
)

type ActionData struct {
	Partitions []*ResizePartition `bson:"partitions" json:"partitions"`
	Volumes    []*ResizeVolume    `bson:"volumes" json:"volumes"`
}

type ResizePartition struct {
	Number int   `bson:"number" json:"number"`
	Size   int64 `bson:"size" json:"size"`
	Fill   bool  `bson:"fill" json:"fill"`
}

type ResizeVolume struct {
	VolumeGroup string `bson:"volume_group" json:"volume_group"`
	Name        string `bson:"name" json:"name"`
	Size        int64  `bson:"size" json:"size"`
	Fill        bool   `bson:"fill" json:"fill"`
}

func IsManagedVgName(name string) bool {
	return strings.HasPrefix(name, VgNamePrefix)
}

func alignSize(size int64) int64 {
	return size - size%ResizeAlign
}

type resizeFill struct {
	request *ResizeVolume
	volume  *LogicalVolume
}

func (d *Disk) ValidateResize(data *ActionData) (
	errData *errortypes.ErrorData) {

	if d.Metadata == nil || len(d.Metadata.Partitions) == 0 {
		errData = &errortypes.ErrorData{
			Error:   "disk_not_scanned",
			Message: "Disk file system has not been scanned",
		}
		return
	}

	partitions := append([]*Partition{}, d.Metadata.Partitions...)
	sort.SliceStable(partitions, func(i, j int) bool {
		return partitions[i].Start < partitions[j].Start
	})

	byNumber := map[int]*Partition{}
	available := map[int]int64{}
	vgs := map[string]*VolumeGroup{}
	for i, part := range partitions {
		byNumber[part.Number] = part

		limit := d.Metadata.Size
		if i+1 < len(partitions) {
			limit = partitions[i+1].Start
		} else if d.Metadata.Table == "gpt" {
			limit -= gptReserve
		}
		available[part.Number] = alignSize(limit - part.Start)

		if part.VolumeGroup != nil {
			vgs[part.VolumeGroup.Name] = part.VolumeGroup
		}
	}

	pvDelta := map[string]int64{}
	seen := set.NewSet()
	resizePartitions := []*ResizePartition{}
	for _, rp := range data.Partitions {
		if rp == nil {
			continue
		}

		part := byNumber[rp.Number]
		if rp.Number <= 0 || part == nil {
			errData = &errortypes.ErrorData{
				Error:   "partition_not_found",
				Message: fmt.Sprintf("Partition %d not found", rp.Number),
			}
			return
		}

		if seen.Contains(rp.Number) {
			errData = &errortypes.ErrorData{
				Error:   "partition_duplicate",
				Message: fmt.Sprintf("Partition %d repeated", rp.Number),
			}
			return
		}
		seen.Add(rp.Number)

		switch part.FileSystem {
		case Xfs, Ext4:
			break
		case Lvm:
			if part.VolumeGroup == nil ||
				!IsManagedVgName(part.VolumeGroup.Name) {

				errData = &errortypes.ErrorData{
					Error:   "volume_group_unsupported",
					Message: "Volume group cannot be resized",
				}
				return
			}
			break
		default:
			errData = &errortypes.ErrorData{
				Error: "file_system_unsupported",
				Message: "Only xfs, ext4 and lvm partitions " +
					"can be resized",
			}
			return
		}

		if rp.Fill {
			rp.Size = available[rp.Number]
		} else {
			rp.Size = alignSize(rp.Size)
		}

		if rp.Size < ResizeMinSize {
			errData = &errortypes.ErrorData{
				Error:   "size_invalid",
				Message: "Partition size too small",
			}
			return
		}

		if rp.Size > available[rp.Number] {
			errData = &errortypes.ErrorData{
				Error:   "size_invalid",
				Message: "Partition size exceeds available space",
			}
			return
		}

		if rp.Size < part.Size && part.FileSystem == Xfs {
			errData = &errortypes.ErrorData{
				Error:   "xfs_shrink",
				Message: "XFS partitions cannot be shrunk",
			}
			return
		}

		if rp.Size == part.Size {
			continue
		}

		if part.FileSystem == Lvm {
			pvDelta[part.VolumeGroup.Name] = rp.Size - part.Size
		}

		resizePartitions = append(resizePartitions, rp)
	}

	lvDelta := map[string]int64{}
	fills := map[string]*resizeFill{}
	seenVolumes := set.NewSet()
	resizeVolumes := []*ResizeVolume{}
	for _, rv := range data.Volumes {
		if rv == nil {
			continue
		}

		vg := vgs[rv.VolumeGroup]
		if vg == nil {
			errData = &errortypes.ErrorData{
				Error: "volume_group_not_found",
				Message: fmt.Sprintf("Volume group %s not found",
					rv.VolumeGroup),
			}
			return
		}

		if !IsManagedVgName(vg.Name) {
			errData = &errortypes.ErrorData{
				Error:   "volume_group_unsupported",
				Message: "Volume group cannot be resized",
			}
			return
		}

		var lv *LogicalVolume
		for _, volume := range vg.Volumes {
			if volume.Name == rv.Name {
				lv = volume
				break
			}
		}
		if lv == nil {
			errData = &errortypes.ErrorData{
				Error:   "volume_not_found",
				Message: fmt.Sprintf("Volume %s not found", rv.Name),
			}
			return
		}

		key := vg.Name + "/" + lv.Name
		if seenVolumes.Contains(key) {
			errData = &errortypes.ErrorData{
				Error:   "volume_duplicate",
				Message: fmt.Sprintf("Volume %s repeated", key),
			}
			return
		}
		seenVolumes.Add(key)

		switch lv.FileSystem {
		case Xfs, Ext4:
			break
		default:
			errData = &errortypes.ErrorData{
				Error:   "file_system_unsupported",
				Message: "Only xfs and ext4 volumes can be resized",
			}
			return
		}

		if rv.Fill {
			if fills[vg.Name] != nil {
				errData = &errortypes.ErrorData{
					Error: "volume_fill_multiple",
					Message: "Only one volume per volume group can " +
						"use all free space",
				}
				return
			}
			fills[vg.Name] = &resizeFill{
				request: rv,
				volume:  lv,
			}
			continue
		}

		rv.Size = alignSize(rv.Size)

		if rv.Size < ResizeMinSize {
			errData = &errortypes.ErrorData{
				Error:   "size_invalid",
				Message: "Volume size too small",
			}
			return
		}

		if rv.Size < lv.Size && lv.FileSystem == Xfs {
			errData = &errortypes.ErrorData{
				Error:   "xfs_shrink",
				Message: "XFS volumes cannot be shrunk",
			}
			return
		}

		if rv.Size == lv.Size {
			continue
		}

		lvDelta[vg.Name] += rv.Size - lv.Size
		resizeVolumes = append(resizeVolumes, rv)
	}

	for vgName, vg := range vgs {
		free := vg.Free + pvDelta[vgName] - lvDelta[vgName]
		if free < 0 {
			errData = &errortypes.ErrorData{
				Error: "size_invalid",
				Message: fmt.Sprintf(
					"Volume sizes exceed volume group %s", vgName),
			}
			return
		}

		fill := fills[vgName]
		if fill != nil {
			free = alignSize(free)
			if free > 0 {
				fill.request.Size = fill.volume.Size + free
				resizeVolumes = append(resizeVolumes, fill.request)
			}
		}
	}

	if len(resizePartitions) == 0 && len(resizeVolumes) == 0 {
		errData = &errortypes.ErrorData{
			Error:   "resize_empty",
			Message: "No partition or volume size changes",
		}
		return
	}

	data.Partitions = resizePartitions
	data.Volumes = resizeVolumes

	return
}

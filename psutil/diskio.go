package psutil

import (
	"github.com/pritunl/pritunl-cloud/metric"
)

type diskIoStat struct {
	Name       string
	BytesRead  uint64
	BytesWrite uint64
	CountRead  uint64
	CountWrite uint64
	TimeRead   uint64
	TimeWrite  uint64
	TimeIo     uint64
}

var diskIoPrev = map[string]*diskIoStat{}

func GetDiskIo() (disks []*metric.DiskIoDisk, err error) {
	stats, err := diskIoList()
	if err != nil {
		return
	}

	seen := map[string]bool{}
	for _, stat := range stats {
		seen[stat.Name] = true

		prev := diskIoPrev[stat.Name]
		diskIoPrev[stat.Name] = stat
		if prev == nil {
			continue
		}

		disks = append(disks, &metric.DiskIoDisk{
			Node:       stat.Name,
			BytesRead:  delta(stat.BytesRead, prev.BytesRead),
			BytesWrite: delta(stat.BytesWrite, prev.BytesWrite),
			CountRead:  delta(stat.CountRead, prev.CountRead),
			CountWrite: delta(stat.CountWrite, prev.CountWrite),
			TimeRead:   delta(stat.TimeRead, prev.TimeRead),
			TimeWrite:  delta(stat.TimeWrite, prev.TimeWrite),
			TimeIo:     delta(stat.TimeIo, prev.TimeIo),
		})
	}

	for name := range diskIoPrev {
		if !seen[name] {
			delete(diskIoPrev, name)
		}
	}

	return
}

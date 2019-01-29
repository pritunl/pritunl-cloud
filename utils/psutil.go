package utils

import (
	"runtime"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
)

func MemoryUsed() (used, total float64, err error) {
	virt, err := mem.VirtualMemory()
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrapf(err, "utils: Failed to read virtual memory"),
		}
		return
	}

	used = ToFixed(virt.UsedPercent, 2)
	total = ToFixed(float64(virt.Total)/float64(1073741824), 2)

	return
}

type LoadStat struct {
	CpuUnits int
	Load1    float64
	Load5    float64
	Load15   float64
}

func LoadAverage() (ld *LoadStat, err error) {
	count := runtime.NumCPU()
	countFloat := float64(count)

	avg, err := load.Avg()
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrapf(err, "utils: Failed to read load average"),
		}
		return
	}

	ld = &LoadStat{
		CpuUnits: count,
		Load1:    ToFixed(avg.Load1/countFloat*100, 2),
		Load5:    ToFixed(avg.Load5/countFloat*100, 2),
		Load15:   ToFixed(avg.Load15/countFloat*100, 2),
	}

	return
}

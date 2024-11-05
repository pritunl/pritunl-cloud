package utils

import (
	"runtime"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"golang.org/x/sys/unix"
)

type MemInfo struct {
	Total                uint64
	Free                 uint64
	Available            uint64
	Buffers              uint64
	Cached               uint64
	Used                 uint64
	UsedPercent          float64
	Dirty                uint64
	SwapTotal            uint64
	SwapFree             uint64
	SwapUsed             uint64
	SwapUsedPercent      float64
	HugePagesTotal       uint64
	HugePagesFree        uint64
	HugePagesReserved    uint64
	HugePagesUsed        uint64
	HugePagesUsedPercent float64
	HugePageSize         uint64
}

func GetMemInfo() (info *MemInfo, err error) {
	info = &MemInfo{}

	totalMem, err := unix.SysctlUint64("hw.physmem")
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "utils: Failed to read physmem"),
		}
		return
	}
	info.Total = totalMem / 1024

	pageSize, err := unix.SysctlUint64("hw.pagesize")
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "utils: Failed to read pagesize"),
		}
		return
	}

	freePages, err := unix.SysctlUint64("vm.stats.vm.v_free_count")
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "utils: Failed to read freecount"),
		}
		return
	}
	info.Free = (freePages * pageSize) / 1024

	info.Available = info.Free

	info.Used = info.Total - info.Free
	if info.Total > 0 {
		info.UsedPercent = float64(info.Used) / float64(info.Total) * 100.0
	}

	info.HugePageSize = pageSize / 1024

	return info, nil
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
	loads := make([]float64, 3)

	n, err := unix.Getloadavg(loads)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "utils: Failed to read loadavg"),
		}
		return
	}
	if n < 3 {
		err = &errortypes.ParseError{
			errors.Wrap(err, "utils: Failed to parse loadavg"),
		}
		return
	}

	load1, load5, load15 := loads[0], loads[1], loads[2]

	ld = &LoadStat{
		CpuUnits: count,
		Load1:    ToFixed(load1/countFloat*100, 2),
		Load5:    ToFixed(load5/countFloat*100, 2),
		Load15:   ToFixed(load15/countFloat*100, 2),
	}

	return
}

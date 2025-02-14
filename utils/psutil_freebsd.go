package utils

import (
	"encoding/binary"
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

func getSysctlUint64(name string) (uint64, error) {
	value32, err := unix.SysctlUint32(name)
	if err == nil {
		return uint64(value32), nil
	}
	return unix.SysctlUint64(name)
}

func GetMemInfo() (info *MemInfo, err error) {
	info = &MemInfo{}

	totalMem, err := getSysctlUint64("hw.physmem")
	if err != nil {
		return nil, &errortypes.ReadError{
			errors.Wrap(err, "utils: Failed to read physmem"),
		}
	}
	info.Total = totalMem / 1024

	pageSize, err := getSysctlUint64("hw.pagesize")
	if err != nil {
		return nil, &errortypes.ReadError{
			errors.Wrap(err, "utils: Failed to read pagesize"),
		}
	}

	freePages, err := getSysctlUint64("vm.stats.vm.v_free_count")
	if err != nil {
		return nil, &errortypes.ReadError{
			errors.Wrap(err, "utils: Failed to read freecount"),
		}
	}

	info.Free = (uint64(freePages) * uint64(pageSize)) / 1024
	info.Available = info.Free
	info.Used = info.Total - info.Free

	if info.Total > 0 {
		info.UsedPercent = float64(info.Used) / float64(info.Total) * 100.0
	}

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

	loadavgRaw, err := unix.SysctlRaw("vm.loadavg")
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "utils: Failed to read loadavg"),
		}
		return
	}

	if len(loadavgRaw) < 12 {
		err = &errortypes.ReadError{
			errors.New("utils: Invalid loadavg size"),
		}
		return
	}

	const fscale = 1 << 16
	load1 := float64(binary.LittleEndian.Uint32(loadavgRaw[0:4])) / fscale
	load5 := float64(binary.LittleEndian.Uint32(loadavgRaw[4:8])) / fscale
	load15 := float64(binary.LittleEndian.Uint32(loadavgRaw[8:12])) / fscale

	ld = &LoadStat{
		CpuUnits: count,
		Load1:    ToFixed(load1/countFloat*100, 2),
		Load5:    ToFixed(load5/countFloat*100, 2),
		Load15:   ToFixed(load15/countFloat*100, 2),
	}

	return
}

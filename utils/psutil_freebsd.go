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

	pageSize, err := getSysctlUint64("hw.pagesize")
	if err != nil {
		return nil, &errortypes.ReadError{
			errors.Wrap(err, "utils: Failed to read pagesize"),
		}
	}

	physmem, err := getSysctlUint64("hw.physmem")
	if err != nil {
		return nil, &errortypes.ReadError{
			errors.Wrap(err, "utils: Failed to read physmem"),
		}
	}

	freePages, err := getSysctlUint64("vm.stats.vm.v_free_count")
	if err != nil {
		return nil, &errortypes.ReadError{
			errors.Wrap(err, "utils: Failed to read free count"),
		}
	}

	activePages, err := getSysctlUint64("vm.stats.vm.v_active_count")
	if err != nil {
		return nil, &errortypes.ReadError{
			errors.Wrap(err, "utils: Failed to read active count"),
		}
	}

	inactivePages, err := getSysctlUint64("vm.stats.vm.v_inactive_count")
	if err != nil {
		return nil, &errortypes.ReadError{
			errors.Wrap(err, "utils: Failed to read inactive count"),
		}
	}

	wiredPages, err := getSysctlUint64("vm.stats.vm.v_wire_count")
	if err != nil {
		return nil, &errortypes.ReadError{
			errors.Wrap(err, "utils: Failed to read wired count"),
		}
	}

	laundryPages, _ := getSysctlUint64("vm.stats.vm.v_laundry_count")
	cachePages, _ := getSysctlUint64("vm.stats.vm.v_cache_count")

	availablePages := freePages + inactivePages + laundryPages + cachePages
	usedPages := activePages + wiredPages

	info.Total = physmem / 1024
	info.Free = freePages * pageSize / 1024
	info.Available = availablePages * pageSize / 1024
	info.Cached = (inactivePages + laundryPages + cachePages) *
		pageSize / 1024
	info.Used = usedPages * pageSize / 1024

	if physmem > 0 {
		info.UsedPercent = float64(usedPages*pageSize) /
			float64(physmem) * 100.0
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

	fscale := float64(1 << 11)
	if len(loadavgRaw) >= 20 {
		readFscale := float64(binary.LittleEndian.Uint32(loadavgRaw[16:20]))
		if readFscale > 0 {
			fscale = readFscale
		}
	}

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

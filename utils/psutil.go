package utils

import (
	"runtime"
	"strconv"
	"strings"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
)

type MemInfo struct {
	Total             uint64
	Free              uint64
	Available         uint64
	Buffers           uint64
	Cached            uint64
	Used              uint64
	UsedPercent       float64
	Dirty             uint64
	SwapTotal         uint64
	SwapFree          uint64
	SwapUsed          uint64
	SwapUsedPercent   float64
	HugePagesTotal    uint64
	HugePagesFree     uint64
	HugePagesReserved uint64
	HugePageSize      uint64
}

func GetMemInfo() (info *MemInfo, err error) {
	info = &MemInfo{}

	lines, err := ReadLines("/proc/meminfo")
	if err != nil {
		return
	}

	for _, line := range lines {
		fields := strings.Split(line, ":")
		if len(fields) != 2 {
			continue
		}
		key := strings.TrimSpace(fields[0])
		value := strings.TrimSpace(fields[1])
		value = strings.Replace(value, " kB", "", -1)

		switch key {
		case "MemTotal":
			valueInt, e := strconv.ParseUint(value, 10, 64)
			if e != nil {
				err = &errortypes.ParseError{
					errors.Wrap(e, "utils: Failed to parse mem total"),
				}
				return
			}
			info.Total = valueInt
		case "MemFree":
			valueInt, e := strconv.ParseUint(value, 10, 64)
			if e != nil {
				err = &errortypes.ParseError{
					errors.Wrap(e, "utils: Failed to parse mem free"),
				}
				return
			}
			info.Free = valueInt
		case "MemAvailable":
			valueInt, e := strconv.ParseUint(value, 10, 64)
			if e != nil {
				err = &errortypes.ParseError{
					errors.Wrap(e, "utils: Failed to parse mem available"),
				}
				return
			}
			info.Available = valueInt
		case "Buffers":
			valueInt, e := strconv.ParseUint(value, 10, 64)
			if e != nil {
				err = &errortypes.ParseError{
					errors.Wrap(e, "utils: Failed to parse buffers"),
				}
				return
			}
			info.Buffers = valueInt
		case "Cached":
			valueInt, e := strconv.ParseUint(value, 10, 64)
			if e != nil {
				err = &errortypes.ParseError{
					errors.Wrap(e, "utils: Failed to parse cached"),
				}
				return
			}
			info.Cached = valueInt
		case "Dirty":
			valueInt, e := strconv.ParseUint(value, 10, 64)
			if e != nil {
				err = &errortypes.ParseError{
					errors.Wrap(e, "utils: Failed to parse dirty"),
				}
				return
			}
			info.Dirty = valueInt
		case "SwapTotal":
			valueInt, e := strconv.ParseUint(value, 10, 64)
			if e != nil {
				err = &errortypes.ParseError{
					errors.Wrap(e, "utils: Failed to parse swap total"),
				}
				return
			}
			info.SwapTotal = valueInt
		case "SwapFree":
			valueInt, e := strconv.ParseUint(value, 10, 64)
			if e != nil {
				err = &errortypes.ParseError{
					errors.Wrap(e, "utils: Failed to parse swap free"),
				}
				return
			}
			info.SwapFree = valueInt
		case "HugePages_Total":
			valueInt, e := strconv.ParseUint(value, 10, 64)
			if e != nil {
				err = &errortypes.ParseError{
					errors.Wrap(e, "utils: Failed to parse hugepages total"),
				}
				return
			}
			info.HugePagesTotal = valueInt
		case "HugePages_Free":
			valueInt, e := strconv.ParseUint(value, 10, 64)
			if e != nil {
				err = &errortypes.ParseError{
					errors.Wrap(e, "utils: Failed to parse hugepages total"),
				}
				return
			}
			info.HugePagesFree = valueInt
		case "HugePages_Rsvd":
			valueInt, e := strconv.ParseUint(value, 10, 64)
			if e != nil {
				err = &errortypes.ParseError{
					errors.Wrap(e,
						"utils: Failed to parse hugepages reserved"),
				}
				return
			}
			info.HugePagesReserved = valueInt
		case "Hugepagesize":
			valueInt, e := strconv.ParseUint(value, 10, 64)
			if e != nil {
				err = &errortypes.ParseError{
					errors.Wrap(e, "utils: Failed to parse hugepages size"),
				}
				return
			}
			info.HugePageSize = valueInt
		}
	}

	info.Used = info.Total - info.Free - info.Buffers - info.Cached
	info.UsedPercent = float64(info.Used) / float64(info.Total) * 100.0

	info.SwapUsed = info.SwapTotal - info.SwapFree
	if info.SwapUsed != 0 {
		info.SwapUsedPercent = float64(
			info.SwapUsed) / float64(info.SwapTotal) * 100.0
	}

	return
}

func MemoryUsed() (used, total float64, err error) {
	virt, err := mem.VirtualMemory()
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "utils: Failed to read virtual memory"),
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
			errors.Wrap(err, "utils: Failed to read load average"),
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

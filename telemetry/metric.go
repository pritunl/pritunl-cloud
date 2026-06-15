package telemetry

import (
	"runtime"
	"time"

	"github.com/pritunl/pritunl-cloud/metric"
	"github.com/pritunl/pritunl-cloud/psutil"
	"github.com/pritunl/pritunl-cloud/utils"
)

const devLimit = 32

var Metrics = &Telemetry[*metric.Sample]{
	RefreshRate: 1 * time.Minute,
	Queue:       360,
	Refresher:   MetricsRefresh,
	Validate: func(data *metric.Sample) *metric.Sample {
		if data.Disk != nil && len(data.Disk.Mounts) > devLimit {
			data.Disk.Mounts = data.Disk.Mounts[:devLimit]
		}
		if data.DiskIo != nil && len(data.DiskIo.Disks) > devLimit {
			data.DiskIo.Disks = data.DiskIo.Disks[:devLimit]
		}
		if data.Network != nil &&
			len(data.Network.Interfaces) > devLimit {

			data.Network.Interfaces =
				data.Network.Interfaces[:devLimit]
		}
		return data
	},
}

func MetricsRefresh() (sample *metric.Sample, err error) {
	timestamp := time.Now().UTC().Truncate(1 * time.Minute)

	sample = &metric.Sample{
		Timestamp: timestamp,
	}

	sys := &metric.System{
		Timestamp: timestamp,
		CpuCores:  runtime.NumCPU(),
	}

	if cpu, e := psutil.GetCpu(); e == nil {
		sys.CpuUsage = cpu
	}
	if procs, e := psutil.GetProcesses(); e == nil {
		sys.Processes = procs
	}

	mem, err := utils.GetMemInfo()
	if err == nil {
		sys.MemUsage = utils.ToFixed(mem.UsedPercent, 2)
		sys.MemTotal = int(mem.Total / 1024)
		sys.SwapUsage = utils.ToFixed(mem.SwapUsedPercent, 2)
		sys.SwapTotal = int(mem.SwapTotal / 1024)
		sys.HugeUsage = utils.ToFixed(mem.HugePagesUsedPercent, 2)
		sys.HugeTotal = int(mem.HugePagesTotal * mem.HugePageSize / 1024)
	}
	sample.System = sys

	load, err := utils.LoadAverage()
	if err == nil {
		sample.Load = &metric.Load{
			Timestamp: timestamp,
			Load1:     load.Load1,
			Load5:     load.Load5,
			Load15:    load.Load15,
		}
	}

	mounts, err := psutil.GetDisks()
	if err == nil && len(mounts) > 0 {
		sample.Disk = &metric.Disk{
			Timestamp: timestamp,
			Mounts:    mounts,
		}
	}

	disks, err := psutil.GetDiskIo()
	if err == nil && len(disks) > 0 {
		sample.DiskIo = &metric.DiskIo{
			Timestamp: timestamp,
			Disks:     disks,
		}
	}

	ifaces, err := psutil.GetNetwork()
	if err == nil && len(ifaces) > 0 {
		sample.Network = &metric.Network{
			Timestamp:  timestamp,
			Interfaces: ifaces,
		}
	}

	return
}

func init() {
	Register(Metrics)
}

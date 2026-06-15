package psutil

import (
	"github.com/pritunl/pritunl-cloud/metric"
)

type networkStat struct {
	Name        string
	BytesSent   uint64
	BytesRecv   uint64
	PacketsSent uint64
	PacketsRecv uint64
	ErrorsSent  uint64
	ErrorsRecv  uint64
	DropsSent   uint64
	DropsRecv   uint64
	FifoSent    uint64
	FifoRecv    uint64
}

var networkPrev = map[string]*networkStat{}

func GetNetwork() (ifaces []*metric.Interface, err error) {
	stats, err := networkList()
	if err != nil {
		return
	}

	seen := map[string]bool{}
	for _, stat := range stats {
		seen[stat.Name] = true

		prev := networkPrev[stat.Name]
		networkPrev[stat.Name] = stat
		if prev == nil {
			continue
		}

		ifaces = append(ifaces, &metric.Interface{
			Name:        stat.Name,
			BytesSent:   delta(stat.BytesSent, prev.BytesSent),
			BytesRecv:   delta(stat.BytesRecv, prev.BytesRecv),
			PacketsSent: delta(stat.PacketsSent, prev.PacketsSent),
			PacketsRecv: delta(stat.PacketsRecv, prev.PacketsRecv),
			ErrorsSent:  delta(stat.ErrorsSent, prev.ErrorsSent),
			ErrorsRecv:  delta(stat.ErrorsRecv, prev.ErrorsRecv),
			DropsSent:   delta(stat.DropsSent, prev.DropsSent),
			DropsRecv:   delta(stat.DropsRecv, prev.DropsRecv),
			FifoSent:    delta(stat.FifoSent, prev.FifoSent),
			FifoRecv:    delta(stat.FifoRecv, prev.FifoRecv),
		})
	}

	for name := range networkPrev {
		if !seen[name] {
			delete(networkPrev, name)
		}
	}

	return
}

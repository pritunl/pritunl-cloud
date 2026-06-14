package psutil

import (
	"github.com/pritunl/pritunl-cloud/utils"
)

var (
	cpuPrevTotal uint64
	cpuPrevIdle  uint64
)

func GetCpu() (usage float64, err error) {
	total, idle, err := cpuTimes()
	if err != nil {
		return
	}

	if cpuPrevTotal != 0 && total > cpuPrevTotal {
		totalDelta := total - cpuPrevTotal
		idleDelta := delta(idle, cpuPrevIdle)
		if totalDelta > 0 {
			usage = utils.ToFixed(
				float64(totalDelta-idleDelta)/float64(totalDelta)*100, 2)
		}
	}

	cpuPrevTotal = total
	cpuPrevIdle = idle

	return
}

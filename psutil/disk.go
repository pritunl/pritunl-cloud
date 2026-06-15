package psutil

import (
	"github.com/pritunl/pritunl-cloud/metric"
)

const diskMinSize = 50 * 1024 * 1024

func GetDisks() (mounts []*metric.Mount, err error) {
	mounts, err = disksList()
	if err != nil {
		return
	}

	return
}

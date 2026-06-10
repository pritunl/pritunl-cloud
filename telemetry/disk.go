package telemetry

import (
	"time"
)

const diskMinSize = 50 * 1024 * 1024

var Disks = &Telemetry[[]*Disk]{
	TransmitRate: 3 * time.Minute,
	RefreshRate:  10 * time.Minute,
	Refresher:    DisksRefresh,
	Validate: func(data []*Disk) []*Disk {
		if len(data) > 50 {
			return data[:50]
		}
		return data
	},
}

type Disk struct {
	Mount string `bson:"mount" json:"mount"`
	Used  int64  `bson:"used" json:"used"`
	Size  int64  `bson:"size" json:"size"`
}

func DisksRefresh() (disks []*Disk, err error) {
	disks, err = disksList()
	if err != nil {
		return
	}
	if disks == nil {
		disks = []*Disk{}
	}
	return
}

func init() {
	Register(Disks)
}

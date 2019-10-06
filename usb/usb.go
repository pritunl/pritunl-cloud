package usb

import (
	"strings"
	"sync"
	"time"

	"github.com/pritunl/pritunl-cloud/utils"
)

var (
	syncLast    time.Time
	syncLock    sync.Mutex
	syncDevices []*Device
)

type Device struct {
	Name    string `bson:"name" json:"name"`
	Vendor  string `bson:"vendor" json:"vendor"`
	Product string `bson:"product" json:"product"`
}

func GetDevices() (devices []*Device, err error) {
	if time.Since(syncLast) < 30*time.Second {
		devices = syncDevices
		return
	}
	syncLock.Lock()
	defer syncLock.Unlock()

	output, err := utils.ExecCombinedOutput("", "lsusb")
	if err != nil {
		return
	}

	outputLines := strings.Split(output, "\n")
	for _, line := range outputLines {
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		if fields[4] != "ID" {
			continue
		}

		deviceId := strings.SplitN(fields[5], ":", 2)
		if len(deviceId) != 2 {
			continue
		}

		device := &Device{
			Name:    strings.Join(fields[6:], " "),
			Vendor:  deviceId[0],
			Product: deviceId[1],
		}
		devices = append(devices, device)
	}

	syncDevices = devices
	syncLast = time.Now()

	return
}

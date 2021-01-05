package pci

import (
	"strings"
	"sync"
	"time"

	"github.com/pritunl/pritunl-cloud/utils"
)

var (
	syncLast  time.Time
	syncLock  sync.Mutex
	syncCache []*Device
)

type Device struct {
	Slot   string `bson:"slot" json:"slot"`
	Class  string `bson:"class" json:"class"`
	Name   string `bson:"name" json:"name"`
	Driver string `bson:"driver" json:"driver"`
}

func GetVfio() (devices []*Device, err error) {
	if time.Since(syncLast) < 30*time.Second {
		devices = syncCache
		return
	}
	syncLock.Lock()
	defer syncLock.Unlock()

	devices = []*Device{}

	output, err := utils.ExecOutput("", "lspci", "-v")
	if err != nil {
		return
	}

	dev := &Device{}

	outputLines := strings.Split(output, "\n")
	for _, line := range outputLines {
		if strings.TrimSpace(line) == "" {
			if dev.Slot != "" && dev.Name != "" && dev.Driver == "vfio-pci" {
				devices = append(devices, dev)
			}

			dev = &Device{}

			continue
		}

		if dev.Slot == "" {
			lines := strings.SplitN(line, " ", 2)
			if len(lines) != 2 {
				continue
			}

			names := strings.SplitN(lines[1], ":", 2)
			if len(names) != 2 {
				continue
			}

			dev.Slot = strings.TrimSpace(lines[0])
			dev.Class = strings.TrimSpace(names[0])
			dev.Name = strings.TrimSpace(names[1])
		} else if strings.Contains(line, "Kernel driver in use:") {
			lines := strings.SplitN(line, ":", 2)
			if len(lines) != 2 {
				continue
			}

			dev.Driver = strings.TrimSpace(lines[1])
		}
	}

	syncCache = devices
	syncLast = time.Now()

	return
}

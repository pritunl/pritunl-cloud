package pci

import (
	"regexp"
	"strings"
	"time"

	"github.com/pritunl/pritunl-cloud/utils"
)

var (
	reg = regexp.MustCompile(
		"[a-fA-F0-9][a-fA-F0-9]:[a-fA-F0-9][a-fA-F0-9].[0-9]")
)

func CheckSlot(slot string) bool {
	return reg.MatchString(slot)
}

func GetVfio(slot string) (dev *Device, err error) {
	devices, err := GetVfioAll()
	if err != nil {
		return
	}

	for _, device := range devices {
		if device.Slot == slot {
			dev = device
			return
		}
	}

	return
}

func GetVfioAll() (devices []*Device, err error) {
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
			if dev.Slot != "" && dev.Name != "" &&
				dev.Driver == "vfio-pci" && CheckSlot(dev.Slot) {

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

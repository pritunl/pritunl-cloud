package systemd

import (
	"github.com/pritunl/pritunl-cloud/utils"
	"strings"
	"sync"
)

var (
	systemdLock = sync.RWMutex{}
)

func Reload() (err error) {
	systemdLock.Lock()
	defer systemdLock.Unlock()

	err = utils.Exec("", "systemctl", "daemon-reload")
	if err != nil {
		return
	}

	return
}

func Start(unit string) (err error) {
	systemdLock.RLock()
	defer systemdLock.RUnlock()

	err = utils.Exec("", "systemctl", "start", unit)
	if err != nil {
		return
	}

	return
}

func Stop(unit string) (err error) {
	systemdLock.RLock()
	defer systemdLock.RUnlock()

	err = utils.Exec("", "systemctl", "stop", unit)
	if err != nil {
		return
	}

	return
}

func GetState(unit string) (state string, err error) {
	systemdLock.RLock()
	defer systemdLock.RUnlock()

	output, _ := utils.ExecOutput("", "systemctl", "is-active", unit)
	state = strings.TrimSpace(output)

	return
}

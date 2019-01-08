package systemd

import (
	"github.com/pritunl/pritunl-cloud/utils"
	"strings"
	"sync"
	"time"
)

var (
	systemdLock = sync.Mutex{}
)

func Reload() (err error) {
	systemdLock.Lock()
	defer systemdLock.Unlock()

	err = utils.Exec("", "systemctl", "daemon-reload")
	if err != nil {
		return
	}

	time.Sleep(100 * time.Millisecond)

	return
}

func Start(unit string) (err error) {
	systemdLock.Lock()
	defer systemdLock.Unlock()

	err = utils.Exec("", "systemctl", "start", unit)
	if err != nil {
		return
	}

	time.Sleep(300 * time.Millisecond)

	return
}

func Stop(unit string) (err error) {
	systemdLock.Lock()
	defer systemdLock.Unlock()

	err = utils.Exec("", "systemctl", "stop", unit)
	if err != nil {
		return
	}

	time.Sleep(300 * time.Millisecond)

	return
}

func GetState(unit string) (state string, err error) {
	systemdLock.Lock()
	defer systemdLock.Unlock()

	output, _ := utils.ExecOutput("", "systemctl", "is-active", unit)
	state = strings.TrimSpace(output)

	return
}

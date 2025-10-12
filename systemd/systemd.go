package systemd

import (
	"strings"
	"sync"
	"time"

	"github.com/pritunl/pritunl-cloud/utils"
)

var (
	systemdLock = sync.Mutex{}
)

func Reload() (err error) {
	systemdLock.Lock()
	defer systemdLock.Unlock()

	_, err = utils.ExecCombinedOutput("", "systemctl", "daemon-reload")
	if err != nil {
		return
	}

	time.Sleep(100 * time.Millisecond)

	return
}

func Start(unit string) (err error) {
	systemdLock.Lock()
	defer systemdLock.Unlock()

	_, err = utils.ExecCombinedOutputLogged(nil, "systemctl", "start", unit)
	if err != nil {
		return
	}

	time.Sleep(300 * time.Millisecond)

	return
}

func Restart(unit string) (err error) {
	systemdLock.Lock()
	defer systemdLock.Unlock()

	_, err = utils.ExecCombinedOutputLogged(nil, "systemctl", "restart", unit)
	if err != nil {
		return
	}

	time.Sleep(300 * time.Millisecond)

	return
}

func Stop(unit string) (err error) {
	systemdLock.Lock()
	defer systemdLock.Unlock()

	_, err = utils.ExecCombinedOutput("", "systemctl", "stop", unit)
	if err != nil {
		return
	}

	time.Sleep(300 * time.Millisecond)

	return
}

func GetState(unit string) (state string, timestamp time.Time, err error) {
	systemdLock.Lock()
	defer systemdLock.Unlock()

	output, _ := utils.ExecOutput("", "systemctl", "show",
		"--no-page", unit)

	timestampStr := ""
	exitCode := ""
	exitStatus := ""

	for _, line := range strings.Split(output, "\n") {
		n := len(line)
		if state == "" && n > 13 && line[:12] == "ActiveState=" {
			state = line[12:]
		} else if exitCode == "" && n > 13 && line[:13] == "ExecMainCode=" {
			exitCode = line[13:]
		} else if exitStatus == "" && n > 15 &&
			line[:15] == "ExecMainStatus=" {

			exitStatus = line[15:]
		} else if timestampStr == "" && n > 24 &&
			line[:23] == "ExecMainStartTimestamp=" {

			timestampStr = line[23:]
		}
	}

	if (state == "failed" && exitCode == "2" && exitStatus == "31") ||
		(state == "failed" && exitCode == "3" && exitStatus == "31") {

		state = "inactive"
	}

	if timestampStr != "" && timestampStr != "0" && timestampStr != "n/a" {
		timestamp, _ = time.Parse("Mon 2006-01-02 15:04:05 MST", timestampStr)
	}

	return
}

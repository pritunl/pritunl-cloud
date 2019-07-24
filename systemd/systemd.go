package systemd

import (
	"strings"
	"sync"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/utils"
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

func GetState(unit string) (state string, timestamp time.Time, err error) {
	systemdLock.Lock()
	defer systemdLock.Unlock()

	output, _ := utils.ExecOutput("", "systemctl", "show",
		"--no-page", unit)

	timestampStr := ""

	for _, line := range strings.Split(output, "\n") {
		n := len(line)
		if state == "" && n > 13 && line[:12] == "ActiveState=" {
			state = line[12:]
		} else if timestampStr == "" && n > 24 &&
			line[:23] == "ExecMainStartTimestamp=" {

			timestampStr = line[23:]
		}
	}

	if timestampStr != "" {
		timestamp, err = time.Parse(
			"Mon 2006-01-02 15:04:05 MST", timestampStr)
		if err != nil {
			err = &errortypes.ParseError{
				errors.Wrap(err, "systemd: Failed to parse service timestamp"),
			}
			return
		}
	}

	return
}

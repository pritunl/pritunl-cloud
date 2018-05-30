package link

import (
	"fmt"
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/utils"
	"time"
	"strings"
)

var (
	offlineTime time.Time
)

func GetStatus() (status Status, err error) {
	status = Status{}

	// TODO Exit status 3
	output, err := utils.ExecOutput("", "ipsec", "status")
	if err != nil {
		err = nil
		return
	}

	isIkeState := false
	ikeState := ""

	for _, line := range strings.Split(output, "\n") {
		lines := strings.SplitN(line, ":", 2)
		if len(lines) != 2 {
			continue
		}

		isIkeState = strings.HasSuffix(lines[0], "]")

		if isIkeState {
			ikeState = strings.SplitN(
				strings.TrimSpace(lines[1]), " ", 2)[0]
		} else {
			if !strings.Contains(lines[1], "reqid") {
				continue
			}

			connId := strings.SplitN(strings.SplitN(
				lines[0], "{", 2)[0], "-", 2)
			connState := strings.SplitN(
				strings.TrimSpace(lines[1]), ",", 2)[0]

			if len(connId) != 2 {
				continue
			}

			switch ikeState {
			case "ESTABLISHED":
				if connState == "INSTALLED" {
					connState = "connected"
				} else {
					connState = "disconnected"
				}
				break
			case "CONNECTING":
				connState = "connecting"
			default:
				connState = "disconnected"
			}

			if _, ok := status[connId[0]]; !ok {
				status[connId[0]] = map[string]string{}
			}

			if _, ok := status[connId[0]][connId[1]]; !ok {
				status[connId[0]][connId[1]] = connState
			} else if (status[connId[0]][connId[1]] == "disconnected") ||
				(status[connId[0]][connId[1]] == "connecting" &&
					connState == "connected") {

				status[connId[0]][connId[1]] = connState
			}
		}
	}

	return
}

func Update(names set.Set) (resetLinks []string, err error) {
	resetLinks = []string{}

	stats, err := GetStatus()
	if err != nil {
		return
	}

	unknown := set.NewSet()
	for stateId, conns := range stats {
		for connId, connStatus := range conns {
			id := fmt.Sprintf("%s-%s", stateId, connId)

			if connStatus == "connected" {
				if names.Contains(id) {
					names.Remove(id)
				} else {
					unknown.Add(id)
				}
			}
		}
	}

	if names.Len() > 0 {
		if !offlineTime.IsZero() {
			disconnectedTimeout := time.Duration(
				settings.Ipsec.DisconnectedTimeout) * time.Second

			if !settings.Ipsec.DisableDisconnectedRestart {
				if time.Since(offlineTime) > disconnectedTimeout {
					for nameInf := range names.Iter() {
						resetLinks = append(resetLinks, nameInf.(string))
					}
					offlineTime = time.Time{}
				}
			} else {
				offlineTime = time.Time{}
			}
		} else {
			offlineTime = time.Now()
		}
	} else {
		offlineTime = time.Time{}
	}

	return
}

package link

import (
	"fmt"
	"strings"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
)

var (
	offlineTime time.Time
)

func GetStatus(vpcId primitive.ObjectID) (status Status, err error) {
	status = Status{}
	namespace := vm.GetLinkNamespace(vpcId, 0)

	output, err := utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", namespace,
		"ipsec", "status",
	)
	if err != nil {
		err = nil
		return
	}

	for _, line := range strings.Split(output, "\n") {
		lines := strings.SplitN(line, ":", 2)
		if len(lines) != 2 {
			continue
		}

		if !strings.HasSuffix(lines[0], "]") {
			continue
		}

		connId := strings.SplitN(strings.SplitN(lines[0], "[", 2)[0], "-", 2)
		connState := strings.SplitN(
			strings.TrimSpace(lines[1]), " ", 2)[0]

		if len(connId) != 2 {
			continue
		}

		switch connState {
		case "ESTABLISHED":
			connState = "connected"
			break
		case "CONNECTING":
			connState = "connecting"
			break
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

	return
}

func Update(vpcId primitive.ObjectID, names set.Set) (
	resetLinks []string, err error) {

	resetLinks = []string{}

	stats, err := GetStatus(vpcId)
	if err != nil {
		return
	}

	LinkStatusLock.Lock()
	LinkStatus[vpcId] = stats
	LinkStatusLock.Unlock()

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

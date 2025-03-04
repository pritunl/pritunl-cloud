package netconf

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/store"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/tools/commander"
	"github.com/sirupsen/logrus"
)

var (
	dhTimestamps       = map[primitive.ObjectID]time.Time{}
	dhTimestampsLock   = sync.Mutex{}
	dhCleanTimestamp   = time.Now()
	DhTimestampsLoaded = false
)

func (n *NetConf) RestartDhClient(db *database.Database) (err error) {
	err = n.Iface1(db)
	if err != nil {
		return
	}

	err = n.Iface2(db, true)
	if err != nil {
		return
	}

	// pid := ""
	// pidData, _ := ioutil.ReadFile(n.DhcpPidPath)
	// if pidData != nil {
	// 	pid = strings.TrimSpace(string(pidData))
	// }

	// if pid != "" {
	// 	_, _ = utils.ExecCombinedOutput("", "kill", pid)
	// }

	// _ = utils.RemoveAll(n.DhcpPidPath)

	pid := ""
	pidData, _ := os.ReadFile(n.Dhcp6PidPath)
	if pidData != nil {
		pid = strings.TrimSpace(string(pidData))
	}

	if pid != "" {
		_, _ = utils.ExecCombinedOutput("", "kill", pid)
	}

	_ = utils.RemoveAll(n.Dhcp6PidPath)

	// if n.NetworkMode == node.Dhcp {
	// 	_, err = utils.ExecCombinedOutputLogged(
	// 		nil,
	// 		"ip", "netns", "exec", n.Namespace,
	// 		"unshare", "--mount",
	// 		"sh", "-c", fmt.Sprintf(
	// 			"mount -t tmpfs none /etc && dhclient -4 -pf %s -lf %s %s",
	// 			n.DhcpPidPath, n.DhcpLeasePath, n.SpaceExternalIface),
	// 	)
	// 	if err != nil {
	// 		return
	// 	}
	// }

	if n.NetworkMode6 == node.Dhcp || n.NetworkMode6 == node.DhcpSlaac {
		resp, e := commander.Exec(&commander.Opt{
			Name: "ip",
			Args: []string{
				"netns", "exec", n.Namespace,
				"unshare", "--mount",
				"sh", "-c", fmt.Sprintf(
					"mount -t tmpfs none /etc && dhclient -6 -pf %s -lf %s %s",
					n.Dhcp6PidPath, n.Dhcp6LeasePath, n.SpaceExternalIface),
			},
			PipeOut: true,
			PipeErr: true,
		})
		if e != nil {
			if resp != nil {
				logrus.WithFields(resp.Map()).Error(
					"netconf: Failed to start ipv6 dhclient")
			}
			err = e
			return
		}
	}

	SetDhTimestamp(n.Virt.Id, time.Now())

	store.SetAddressExpireMulti(n.Virt.Id, 10*time.Second, 20*time.Second)

	return
}

func getDhTimestamp(instId primitive.ObjectID) (timestamp time.Time) {
	pid := 0
	pidData, _ := os.ReadFile(paths.GetDhcp6PidPath(instId, 0))
	if pidData != nil {
		pid, _ = strconv.Atoi(strings.TrimSpace(string(pidData)))
	}

	timestamp, err := utils.GetProcessTimestamp(pid)
	if err != nil {
		timestamp = time.Time{}
		return
	}

	return
}

func cleanupOldEntries() {
	now := time.Now()

	if now.Sub(dhCleanTimestamp) < time.Hour {
		return
	}

	dhTimestampsLock.Lock()
	defer dhTimestampsLock.Unlock()

	for instId, timestamp := range dhTimestamps {
		if now.Sub(timestamp) > 1*time.Hour {
			delete(dhTimestamps, instId)
		}
	}

	dhCleanTimestamp = now
}

func LoadDhTimestamps(instances []*instance.Instance) (err error) {
	newDhTimestamps := map[primitive.ObjectID]time.Time{}
	for _, inst := range instances {
		if !inst.IsActive() {
			continue
		}

		timestamp := getDhTimestamp(inst.Id)
		newDhTimestamps[inst.Id] = timestamp
	}

	dhTimestampsLock.Lock()
	dhTimestamps = newDhTimestamps
	DhTimestampsLoaded = true
	dhTimestampsLock.Unlock()

	return
}

func GetDhTimestamp(instId primitive.ObjectID) (timestamp time.Time) {
	cleanupOldEntries()

	dhTimestampsLock.Lock()
	timestamp, ok := dhTimestamps[instId]
	dhTimestampsLock.Unlock()
	if !ok {
		timestamp = getDhTimestamp(instId)
		if !timestamp.IsZero() {
			dhTimestampsLock.Lock()
			dhTimestamps[instId] = timestamp
			dhTimestampsLock.Unlock()
		}
	}

	return
}

func SetDhTimestamp(instId primitive.ObjectID, timestamp time.Time) {
	dhTimestampsLock.Lock()
	dhTimestamps[instId] = timestamp
	dhTimestampsLock.Unlock()
}

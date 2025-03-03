package netconf

import (
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/utils"
)

var (
	dhTimestamps       = map[primitive.ObjectID]time.Time{}
	dhTimestampsLock   = sync.Mutex{}
	dhCleanTimestamp   = time.Now()
	DhTimestampsLoaded = false
)

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

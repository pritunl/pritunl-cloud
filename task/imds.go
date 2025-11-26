package task

import (
	"math/rand"
	"sync"
	"time"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/imds"
	"github.com/pritunl/pritunl-cloud/imds/types"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/sirupsen/logrus"
)

var imdsSync = &Task{
	Name: "imds_sync",
	Hours: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12,
		13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23},
	Minutes: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12,
		13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25,
		26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38,
		39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51,
		52, 53, 54, 55, 56, 57, 58, 59},
	Seconds: 3 * time.Second,
	Local:   true,
	Handler: imdsSyncHandler,
}

var (
	failTime = map[bson.ObjectID]failTimeData{}
)

type failTimeData struct {
	timestamp time.Time
	logged    bool
}

func test() {
	test := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}

	for _, val := range test {
		go func() {
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			print(val)
		}()
	}
}

func imdsSyncHandler(db *database.Database) (err error) {
	confs := imds.GetConfigs()
	logTtl := time.Duration(
		settings.Hypervisor.ImdsSyncLogTimeout) * time.Second
	restartTtl := time.Duration(
		settings.Hypervisor.ImdsSyncRestartTimeout) * time.Second

	newFailTime := map[bson.ObjectID]failTimeData{}
	newFailTimeLock := sync.Mutex{}
	waiter := &sync.WaitGroup{}
	for _, conf := range confs {
		if conf.Instance == nil || conf.Instance.NetworkNamespace == "" {
			continue
		}

		waiter.Add(1)
		go func(conf *types.Config) {
			defer waiter.Done()

			err := imds.Sync(db, conf.Instance.NetworkNamespace,
				conf.Instance.Id, conf.Instance.Deployment, conf)
			if err != nil {
				newFailTimeLock.Lock()
				ttlData := failTime[conf.Instance.Id]

				if ttlData.timestamp.IsZero() {
					newFailTime[conf.Instance.Id] = failTimeData{
						timestamp: time.Now(),
					}
				} else if time.Since(ttlData.timestamp) > logTtl &&
					!ttlData.logged {

					logrus.WithFields(logrus.Fields{
						"action":   conf.Instance.Action,
						"instance": conf.Instance.Id.Hex(),
						"error":    err,
					}).Error("task: Failed to sync imds")

					newFailTime[conf.Instance.Id] = failTimeData{
						timestamp: ttlData.timestamp,
						logged:    true,
					}
				} else if time.Since(ttlData.timestamp) > restartTtl {
					logrus.WithFields(logrus.Fields{
						"action":   conf.Instance.Action,
						"instance": conf.Instance.Id.Hex(),
						"error":    err,
					}).Error("task: Failed to sync imds, restarting...")

					e := imds.Restart(conf.Instance.Id)
					if e != nil {
						logrus.WithFields(logrus.Fields{
							"action":   conf.Instance.Action,
							"instance": conf.Instance.Id.Hex(),
							"error":    e,
						}).Error("task: Failed to restart imds")
					}
				} else {
					newFailTime[conf.Instance.Id] = failTime[conf.Instance.Id]
				}
				newFailTimeLock.Unlock()
			}
		}(conf)
	}

	waiter.Wait()

	failTime = newFailTime

	return
}

func init() {
	register(imdsSync)
}

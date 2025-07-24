package task

import (
	"sync"
	"time"

	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/imds"
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
	failTime = map[primitive.ObjectID]time.Time{}
)

func imdsSyncHandler(db *database.Database) (err error) {
	confs := imds.GetConfigs()
	timeout := time.Duration(
		settings.Hypervisor.ImdsSyncLogTimeout) * time.Second

	newFailTime := map[primitive.ObjectID]time.Time{}
	newFailTimeLock := sync.Mutex{}
	waiter := &sync.WaitGroup{}
	for _, conf := range confs {
		if conf.Instance == nil {
			continue
		}

		waiter.Add(1)
		go func() {
			defer waiter.Done()

			err := imds.Sync(db, conf.Instance.NetworkNamespace, conf.Instance.Id,
				conf.Instance.Deployment, conf)
			if err != nil {
				newFailTimeLock.Lock()
				if failTime[conf.Instance.Id].IsZero() {
					newFailTime[conf.Instance.Id] = time.Now()
				} else if time.Since(failTime[conf.Instance.Id]) > timeout {
					logrus.WithFields(logrus.Fields{
						"instance": conf.Instance.Id.Hex(),
						"error":    err,
					}).Error("agent: Failed to sync imds")
				} else {
					newFailTime[conf.Instance.Id] = failTime[conf.Instance.Id]
				}
				newFailTimeLock.Unlock()
			}
		}()
	}

	waiter.Wait()

	failTime = newFailTime

	return
}

func init() {
	register(imdsSync)
}

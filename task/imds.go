package task

import (
	"time"

	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/imds"
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

func imdsSyncHandler(db *database.Database) (err error) {
	confs := imds.GetConfigs()

	for _, conf := range confs {
		instId := conf.Instance.Id

		err := imds.Sync(db, instId, conf)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("agent: Failed to sync imds")
		}
	}

	return
}

func init() {
	register(imdsSync)
}

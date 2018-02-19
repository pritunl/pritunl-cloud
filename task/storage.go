package task

import (
	"github.com/Sirupsen/logrus"
	"github.com/pritunl/pritunl-cloud/data"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/storage"
)

var storageSync = &Task{
	Name: "storage_renew",
	Hours: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
		12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23},
	Mins:       []int{0, 14, 29, 44, 59},
	Handler:    storageSyncHandler,
	RunOnStart: true,
}

func storageSyncHandler(db *database.Database) (err error) {
	stores, err := storage.GetAll(db)
	if err != nil {
		return
	}

	for _, store := range stores {
		err = data.Sync(db, store)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"storage_id":   store.Id,
				"storage_name": store.Name,
				"error":        err,
			}).Error("task: Failed to sync storage")
		}
	}

	return
}

func init() {
	register(storageSync)
}

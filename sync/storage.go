package sync

import (
	"github.com/Sirupsen/logrus"
	"github.com/pritunl/pritunl-cloud/data"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/storage"
	"time"
)

func storageUpdate() (err error) {
	db := database.GetDatabase()
	defer db.Close()

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
			}).Error("sync: Failed to sync storage")
		}
	}

	return
}

func storageRunner() {
	for {
		err := storageUpdate()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("sync: Failed to update storage")
			time.Sleep(3 * time.Second)
			continue
		}

		time.Sleep(10 * time.Minute)
	}
}

func initStorage() {
	go storageRunner()
}

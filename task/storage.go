package task

import (
	"github.com/Sirupsen/logrus"
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/pritunl-cloud/data"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/storage"
	"gopkg.in/mgo.v2/bson"
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
	coll := db.Images()

	imgStoreIdsList := []bson.ObjectId{}
	err = coll.Find(&bson.M{}).Distinct("storage", &imgStoreIdsList)
	if err != nil {
		return
	}

	imgStoreIds := set.NewSet()
	for _, storeId := range imgStoreIdsList {
		imgStoreIds.Add(storeId)
	}

	storeIds := set.NewSet()
	stores, err := storage.GetAll(db)
	if err != nil {
		return
	}

	for _, store := range stores {
		storeIds.Add(store.Id)

		err = data.Sync(db, store)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"storage_id":   store.Id,
				"storage_name": store.Name,
				"error":        err,
			}).Error("task: Failed to sync storage")
		}
	}

	imgStoreIds.Subtract(storeIds)

	remStoreIds := []bson.ObjectId{}
	for storeIdInf := range imgStoreIds.Iter() {
		storeId := storeIdInf.(bson.ObjectId)

		logrus.WithFields(logrus.Fields{
			"storage_id": storeId.Hex(),
		}).Warning("task: Cleaning unknown images")

		remStoreIds = append(remStoreIds, storeId)
	}

	if len(remStoreIds) > 0 {
		_, err = coll.RemoveAll(&bson.M{
			"storage": &bson.M{
				"$in": remStoreIds,
			},
		})
		if err != nil {
			return
		}
	}

	return
}

func init() {
	register(storageSync)
}

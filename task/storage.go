package task

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/data"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/storage"
	"github.com/sirupsen/logrus"
)

var storageSync = &Task{
	Name:    "storage_renew",
	Version: 1,
	Hours: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
		12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23},
	Minutes:    []int{0, 14, 29, 44, 59},
	Handler:    storageSyncHandler,
	RunOnStart: true,
}

func storageSyncHandler(db *database.Database) (err error) {
	coll := db.Images()

	imgStoreIdsList := []primitive.ObjectID{}

	storeIdsInf, err := coll.Distinct(
		db,
		"storage",
		&bson.M{},
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	for _, storeIdInf := range storeIdsInf {
		if storeId, ok := storeIdInf.(primitive.ObjectID); ok {
			imgStoreIdsList = append(imgStoreIdsList, storeId)
		}
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
				"storage_id":   store.Id.Hex(),
				"storage_name": store.Name,
				"error":        err,
			}).Error("task: Failed to sync storage")
		}
	}

	imgStoreIds.Subtract(storeIds)

	remStoreIds := []primitive.ObjectID{}
	for storeIdInf := range imgStoreIds.Iter() {
		storeId := storeIdInf.(primitive.ObjectID)

		logrus.WithFields(logrus.Fields{
			"storage_id": storeId.Hex(),
		}).Warning("task: Cleaning unknown images")

		remStoreIds = append(remStoreIds, storeId)
	}

	if len(remStoreIds) > 0 {
		_, err = coll.DeleteMany(db, &bson.M{
			"storage": &bson.M{
				"$in": remStoreIds,
			},
		})
		if err != nil {
			return
		}
	}

	event.PublishDispatch(db, "image.change")

	return
}

func init() {
	register(storageSync)
}

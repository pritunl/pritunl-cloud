package storage

import (
	"github.com/pritunl/pritunl-cloud/database"
	"gopkg.in/mgo.v2/bson"
)

func Get(db *database.Database, storeId bson.ObjectId) (
	store *Storage, err error) {

	coll := db.Storages()
	store = &Storage{}

	err = coll.FindOneId(storeId, store)
	if err != nil {
		return
	}

	return
}

func GetAll(db *database.Database) (stores []*Storage, err error) {
	coll := db.Storages()
	stores = []*Storage{}

	cursor := coll.Find(bson.M{}).Iter()

	nde := &Storage{}
	for cursor.Next(nde) {
		stores = append(stores, nde)
		nde = &Storage{}
	}

	err = cursor.Close()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func Remove(db *database.Database, storeId bson.ObjectId) (err error) {
	coll := db.Images()

	_, err = coll.RemoveAll(&bson.M{
		"storage": storeId,
	})
	if err != nil {
		return
	}

	coll = db.Storages()

	err = coll.Remove(&bson.M{
		"_id": storeId,
	})
	if err != nil {
		err = database.ParseError(err)
		switch err.(type) {
		case *database.NotFoundError:
			err = nil
		default:
			return
		}
	}

	return
}

package scheduler

import (
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
)

func Get(db *database.Database, schdId primitive.ObjectID) (
	schd *Scheduler, err error) {

	coll := db.Schedulers()
	schd = &Scheduler{}

	err = coll.FindOneId(schdId, schd)
	if err != nil {
		return
	}

	return
}

func GetAll(db *database.Database) (schds []*Scheduler, err error) {
	coll := db.Schedulers()
	schds = []*Scheduler{}

	cursor, err := coll.Find(db, bson.M{})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		schd := &Scheduler{}
		err = cursor.Decode(schd)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		schds = append(schds, schd)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func Remove(db *database.Database, schdId primitive.ObjectID) (err error) {
	coll := db.Schedulers()

	_, err = coll.DeleteOne(db, &bson.M{
		"_id": schdId,
	})
	if err != nil {
		err = database.ParseError(err)
		if _, ok := err.(*database.NotFoundError); ok {
			err = nil
		} else {
			return
		}
	}

	return
}

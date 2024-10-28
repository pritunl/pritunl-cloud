package spec

import (
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/mongo-go-driver/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
)

func New(serviceId, unitId primitive.ObjectID, data string) (spc *Commit) {
	spc = &Commit{
		Unit:      unitId,
		Service:   serviceId,
		Timestamp: time.Now(),
		Data:      data,
	}

	return
}

func Get(db *database.Database, commitId primitive.ObjectID) (
	spc *Commit, err error) {

	coll := db.Specs()
	spc = &Commit{}

	err = coll.FindOneId(commitId, spc)
	if err != nil {
		return
	}

	return
}

func GetAll(db *database.Database, query *bson.M) (
	spcs []*Commit, err error) {

	coll := db.Specs()
	spcs = []*Commit{}

	cursor, err := coll.Find(db, query)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		spc := &Commit{}
		err = cursor.Decode(spc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		spcs = append(spcs, spc)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllIds(db *database.Database) (specIds set.Set, err error) {
	coll := db.Specs()
	specIds = set.NewSet()

	cursor, err := coll.Find(
		db,
		bson.M{},
		&options.FindOptions{
			Projection: bson.M{
				"_id": 1,
			},
		},
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		spc := &Commit{}
		err = cursor.Decode(spc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		specIds.Add(spc.Id)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func Remove(db *database.Database, commitId primitive.ObjectID) (err error) {
	coll := db.Specs()

	_, err = coll.DeleteOne(db, &bson.M{
		"_id": commitId,
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

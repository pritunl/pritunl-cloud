package spec

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/mongo-go-driver/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
)

func New(serviceId, unitId primitive.ObjectID, data string) (spc *Spec) {
	spc = &Spec{
		Id: Hash{
			Unit: unitId,
		},
		Service: serviceId,
		Data:    data,
	}

	return
}

func Get(db *database.Database, hash Hash) (
	spc *Spec, err error) {

	coll := db.Specs()
	spc = &Spec{}

	err = coll.FindOneId(hash, spc)
	if err != nil {
		return
	}

	return
}

func GetAll(db *database.Database, query *bson.M) (
	spcs []*Spec, err error) {

	coll := db.Specs()
	spcs = []*Spec{}

	cursor, err := coll.Find(db, query)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		spc := &Spec{}
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

func GetAllIds(db *database.Database) (hashes set.Set, err error) {
	coll := db.Specs()
	hashes = set.NewSet()

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
		spc := &Spec{}
		err = cursor.Decode(spc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		hashes.Add(spc.Id)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func Remove(db *database.Database, hash Hash) (err error) {
	coll := db.Specs()

	_, err = coll.DeleteOne(db, &bson.M{
		"_id": hash,
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

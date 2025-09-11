package unit

import (
	"time"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/mongo-go-driver/v2/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
)

func Get(db *database.Database, unitId bson.ObjectID) (
	unt *Unit, err error) {

	coll := db.Units()
	unt = &Unit{}

	err = coll.FindOneId(unitId, unt)
	if err != nil {
		return
	}

	return
}

func GetOrg(db *database.Database, orgId, unitId bson.ObjectID) (
	unt *Unit, err error) {

	coll := db.Units()
	unt = &Unit{}

	err = coll.FindOne(db, &bson.M{
		"_id":          unitId,
		"organization": orgId,
	}).Decode(unt)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAll(db *database.Database, query *bson.M) (units []*Unit, err error) {
	coll := db.Units()
	units = []*Unit{}

	cursor, err := coll.Find(db, query)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		unt := &Unit{}
		err = cursor.Decode(unt)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		units = append(units, unt)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllMap(db *database.Database, query *bson.M) (
	unitsMap map[bson.ObjectID]*Unit, err error) {

	coll := db.Units()
	unitsMap = map[bson.ObjectID]*Unit{}

	cursor, err := coll.Find(db, query)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		unt := &Unit{}
		err = cursor.Decode(unt)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		unitsMap[unt.Id] = unt
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func NewSpec(db *database.Database,
	podId, unitId bson.ObjectID) (timestamp time.Time,
	index int, err error) {

	coll := db.Units()

	updateOpts := options.FindOneAndUpdate()
	updateOpts.Projection = &bson.M{
		"spec_index":     1,
		"spec_timestamp": 1,
	}
	updateOpts.SetReturnDocument(options.After)

	unit := &Unit{}

	err = coll.FindOneAndUpdate(db, &bson.M{
		"_id": unitId,
		"pod": podId,
	}, &bson.M{
		"$inc": &bson.M{
			"spec_index": 1,
		},
		"$currentDate": &bson.M{
			"spec_timestamp": true,
		},
	}, updateOpts).Decode(unit)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	index = unit.SpecIndex
	timestamp = unit.SpecTimestamp

	return
}

func Remove(db *database.Database, untId bson.ObjectID) (err error) {
	coll := db.Units()

	_, err = coll.DeleteOne(db, &bson.M{
		"_id": untId,
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

func RemoveOrg(db *database.Database, orgId, untId bson.ObjectID) (
	err error) {

	coll := db.Units()

	_, err = coll.DeleteOne(db, &bson.M{
		"_id":          untId,
		"organization": orgId,
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

func RemoveAll(db *database.Database, query *bson.M) (err error) {
	coll := db.Units()

	_, err = coll.DeleteMany(db, query)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func RemoveMulti(db *database.Database,
	untIds []bson.ObjectID) (err error) {

	coll := db.Units()

	_, err = coll.DeleteMany(db, &bson.M{
		"_id": &bson.M{
			"$in": untIds,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func RemoveMultiOrg(db *database.Database, orgId bson.ObjectID,
	untIds []bson.ObjectID) (err error) {

	coll := db.Units()

	_, err = coll.DeleteMany(db, &bson.M{
		"_id": &bson.M{
			"$in": untIds,
		},
		"organization": orgId,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

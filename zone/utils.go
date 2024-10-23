package zone

import (
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/mongo-go-driver/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
)

func Get(db *database.Database, zoneId primitive.ObjectID) (
	zne *Zone, err error) {

	coll := db.Zones()
	zne = &Zone{}

	err = coll.FindOneId(zoneId, zne)
	if err != nil {
		return
	}

	return
}

func GetOne(db *database.Database, query *bson.M) (zne *Zone, err error) {
	coll := db.Zones()
	zne = &Zone{}

	err = coll.FindOne(db, query).Decode(zne)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAll(db *database.Database) (zones []*Zone, err error) {
	coll := db.Zones()
	zones = []*Zone{}

	cursor, err := coll.Find(
		db,
		&bson.M{},
		&options.FindOptions{
			Sort: &bson.D{
				{"name", 1},
			},
		},
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		zne := &Zone{}
		err = cursor.Decode(zne)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		zones = append(zones, zne)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllDatacenter(db *database.Database, dcId primitive.ObjectID) (
	zones []*Zone, err error) {

	coll := db.Zones()
	zones = []*Zone{}

	cursor, err := coll.Find(
		db,
		&bson.M{
			"datacenter": dcId,
		},
		&options.FindOptions{
			Sort: &bson.D{
				{"name", 1},
			},
		},
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		zne := &Zone{}
		err = cursor.Decode(zne)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		zones = append(zones, zne)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllDatacenters(db *database.Database, dcIds []primitive.ObjectID) (
	zones []*Zone, err error) {

	coll := db.Zones()
	zones = []*Zone{}

	cursor, err := coll.Find(
		db,
		&bson.M{
			"datacenter": &bson.M{
				"$in": dcIds,
			},
		},
		&options.FindOptions{
			Sort: &bson.D{
				{"name", 1},
			},
		},
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		zne := &Zone{}
		err = cursor.Decode(zne)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		zones = append(zones, zne)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func Remove(db *database.Database, zoneId primitive.ObjectID) (err error) {
	coll := db.Zones()

	_, err = coll.DeleteOne(db, &bson.M{
		"_id": zoneId,
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

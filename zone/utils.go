package zone

import (
	"context"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
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

func GetAll(db *database.Database) (zones []*Zone, err error) {
	coll := db.Zones()
	zones = []*Zone{}

	cursor, err := coll.Find(context.Background(), &bson.M{})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
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
		context.Background(),
		&bson.M{
			"datacenter": &bson.M{
				"$in": dcIds,
			},
		},
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
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

	_, err = coll.DeleteOne(context.Background(), &bson.M{
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

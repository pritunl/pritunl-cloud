package datacenter

import (
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
)

func Get(db *database.Database, dcId primitive.ObjectID) (
	dc *Datacenter, err error) {

	coll := db.Datacenters()
	dc = &Datacenter{}

	err = coll.FindOneId(dcId, dc)
	if err != nil {
		return
	}

	return
}

func ExistsOrg(db *database.Database, orgId, dcId primitive.ObjectID) (
	exists bool, err error) {

	coll := db.Datacenters()

	count, err := coll.CountDocuments(db, &bson.M{
		"_id": dcId,
		"$or": []*bson.M{
			&bson.M{
				"match_organizations": false,
			},
			&bson.M{
				"organizations": orgId,
			},
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	if count > 0 {
		exists = true
	}

	return
}

func GetAll(db *database.Database) (dcs []*Datacenter, err error) {
	coll := db.Datacenters()
	dcs = []*Datacenter{}

	cursor, err := coll.Find(db, &bson.M{})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		dc := &Datacenter{}
		err = cursor.Decode(dc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		dcs = append(dcs, dc)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetOne(db *database.Database, query *bson.M) (dc *Datacenter, err error) {
	coll := db.Datacenters()
	dc = &Datacenter{}

	err = coll.FindOne(db, query).Decode(dc)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllNamesOrg(db *database.Database, orgId primitive.ObjectID) (
	dcs []*Datacenter, err error) {

	coll := db.Datacenters()
	dcs = []*Datacenter{}

	cursor, err := coll.Find(db, &bson.M{
		"$or": []*bson.M{
			&bson.M{
				"match_organizations": false,
			},
			&bson.M{
				"organizations": orgId,
			},
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		dc := &Datacenter{}
		err = cursor.Decode(dc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		dcs = append(dcs, dc)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func DistinctOrg(db *database.Database, orgId primitive.ObjectID) (
	ids []primitive.ObjectID, err error) {

	coll := db.Datacenters()
	ids = []primitive.ObjectID{}

	idsInf, err := coll.Distinct(db, "_id", &bson.M{
		"$or": []*bson.M{
			&bson.M{
				"match_organizations": false,
			},
			&bson.M{
				"organizations": orgId,
			},
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	for _, idInf := range idsInf {
		if id, ok := idInf.(primitive.ObjectID); ok {
			ids = append(ids, id)
		}
	}

	return
}

func Remove(db *database.Database, dcId primitive.ObjectID) (err error) {
	coll := db.Datacenters()

	_, err = coll.DeleteOne(db, &bson.M{
		"_id": dcId,
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

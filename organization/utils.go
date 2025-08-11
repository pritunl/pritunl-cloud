package organization

import (
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/mongo-go-driver/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
)

func Get(db *database.Database, dcId primitive.ObjectID) (
	dc *Organization, err error) {

	coll := db.Organizations()
	dc = &Organization{}

	err = coll.FindOneId(dcId, dc)
	if err != nil {
		return
	}

	return
}

func GetAll(db *database.Database, query *bson.M) (
	orgs []*Organization, err error) {

	coll := db.Organizations()
	orgs = []*Organization{}

	cursor, err := coll.Find(
		db,
		query,
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
		org := &Organization{}
		err = cursor.Decode(org)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		orgs = append(orgs, org)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllName(db *database.Database) (orgs []*Organization, err error) {
	coll := db.Organizations()
	orgs = []*Organization{}

	cursor, err := coll.Find(
		db,
		&bson.M{},
		&options.FindOptions{
			Sort: &bson.D{
				{"name", 1},
			},
			Projection: &bson.D{
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
		org := &Organization{}
		err = cursor.Decode(org)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		orgs = append(orgs, org)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllNameRoles(db *database.Database, roles []string) (
	orgs []*Organization, err error) {

	coll := db.Organizations()
	orgs = []*Organization{}

	cursor, err := coll.Find(
		db,
		&bson.M{
			"roles": &bson.M{
				"$in": roles,
			},
		},
		&options.FindOptions{
			Projection: &bson.D{
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
		org := &Organization{}
		err = cursor.Decode(org)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		orgs = append(orgs, org)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func Remove(db *database.Database, dcId primitive.ObjectID) (err error) {
	coll := db.Organizations()

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

func Count(db *database.Database) (count int64, err error) {
	coll := db.Organizations()

	count, err = coll.CountDocuments(db, &bson.M{})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

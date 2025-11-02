package organization

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/mongo-go-driver/v2/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/utils"
)

func Get(db *database.Database, dcId bson.ObjectID) (
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
		options.Find().
			SetSort(bson.D{{"name", 1}}),
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
		options.Find().
			SetSort(bson.D{{"name", 1}}).
			SetProjection(bson.D{{"name", 1}}),
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
		options.Find().
			SetProjection(bson.D{{"name", 1}}),
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

func GetAllPaged(db *database.Database, query *bson.M,
	page, pageCount int64) (orgs []*Organization, count int64, err error) {

	coll := db.Organizations()
	orgs = []*Organization{}

	if len(*query) == 0 {
		count, err = coll.EstimatedDocumentCount(db)
		if err != nil {
			err = database.ParseError(err)
			return
		}
	} else {
		count, err = coll.CountDocuments(db, query)
		if err != nil {
			err = database.ParseError(err)
			return
		}
	}

	if pageCount == 0 {
		pageCount = 20
	}
	maxPage := count / pageCount
	if count == pageCount {
		maxPage = 0
	}
	page = utils.Min64(page, maxPage)
	skip := utils.Min64(page*pageCount, count)

	cursor, err := coll.Find(
		db,
		query,
		options.Find().
			SetSort(bson.D{{"name", 1}}).
			SetSkip(skip).
			SetLimit(pageCount),
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

func Remove(db *database.Database, dcId bson.ObjectID) (err error) {
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

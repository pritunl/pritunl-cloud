package plan

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/mongo-go-driver/v2/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/utils"
)

func Get(db *database.Database, plnId bson.ObjectID) (
	pln *Plan, err error) {

	coll := db.Plans()
	pln = &Plan{}

	err = coll.FindOneId(plnId, pln)
	if err != nil {
		return
	}

	return
}

func GetOrg(db *database.Database, orgId, plnId bson.ObjectID) (
	pln *Plan, err error) {

	coll := db.Plans()
	pln = &Plan{}

	err = coll.FindOne(db, &bson.M{
		"_id":          plnId,
		"organization": orgId,
	}).Decode(pln)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func ExistsOrg(db *database.Database, orgId, plnId bson.ObjectID) (
	exists bool, err error) {

	coll := db.Plans()

	n, err := coll.CountDocuments(db, &bson.M{
		"_id":          plnId,
		"organization": orgId,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	if n > 0 {
		exists = true
	}

	return
}

func GetOne(db *database.Database, query *bson.M) (pln *Plan, err error) {
	coll := db.Plans()
	pln = &Plan{}

	err = coll.FindOne(db, query).Decode(pln)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAll(db *database.Database, query *bson.M) (
	plns []*Plan, err error) {

	coll := db.Plans()
	plns = []*Plan{}

	cursor, err := coll.Find(db, query)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		dmn := &Plan{}
		err = cursor.Decode(dmn)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		plns = append(plns, dmn)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllPaged(db *database.Database, query *bson.M,
	page, pageCount int64) (plns []*Plan, count int64, err error) {

	coll := db.Plans()
	plns = []*Plan{}

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
	defer cursor.Close(db)

	for cursor.Next(db) {
		dmn := &Plan{}
		err = cursor.Decode(dmn)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		plns = append(plns, dmn)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllName(db *database.Database, query *bson.M) (
	plns []*Plan, err error) {

	coll := db.Plans()
	plns = []*Plan{}

	cursor, err := coll.Find(
		db,
		query,
		options.Find().
			SetProjection(bson.D{
				{"name", 1},
				{"organization", 1},
			}),
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		dmn := &Plan{}
		err = cursor.Decode(dmn)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		plns = append(plns, dmn)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func Remove(db *database.Database, plnId bson.ObjectID) (err error) {
	coll := db.Plans()

	_, err = coll.DeleteOne(db, &bson.M{
		"_id": plnId,
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

func RemoveOrg(db *database.Database, orgId, plnId bson.ObjectID) (
	err error) {

	coll := db.Plans()

	_, err = coll.DeleteOne(db, &bson.M{
		"_id":          plnId,
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

func RemoveMulti(db *database.Database, plnIds []bson.ObjectID) (err error) {
	coll := db.Plans()

	_, err = coll.DeleteMany(db, &bson.M{
		"_id": &bson.M{
			"$in": plnIds,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func RemoveMultiOrg(db *database.Database, orgId bson.ObjectID,
	plnIds []bson.ObjectID) (err error) {

	coll := db.Plans()

	_, err = coll.DeleteMany(db, &bson.M{
		"_id": &bson.M{
			"$in": plnIds,
		},
		"organization": orgId,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

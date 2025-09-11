package balancer

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/mongo-go-driver/v2/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/utils"
)

func Get(db *database.Database, balncId bson.ObjectID) (
	balnc *Balancer, err error) {

	coll := db.Balancers()
	balnc = &Balancer{}

	err = coll.FindOneId(balncId, balnc)
	if err != nil {
		return
	}

	return
}

func GetOrg(db *database.Database, orgId, balncId bson.ObjectID) (
	balnc *Balancer, err error) {

	coll := db.Balancers()
	balnc = &Balancer{}

	err = coll.FindOne(db, &bson.M{
		"_id":          balncId,
		"organization": orgId,
	}).Decode(balnc)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAll(db *database.Database, query *bson.M) (
	balncs []*Balancer, err error) {

	coll := db.Balancers()
	balncs = []*Balancer{}

	cursor, err := coll.Find(db, query)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		balnc := &Balancer{}
		err = cursor.Decode(balnc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		balncs = append(balncs, balnc)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllPaged(db *database.Database, query *bson.M,
	page, pageCount int64) (balncs []*Balancer, count int64, err error) {

	coll := db.Balancers()
	balncs = []*Balancer{}

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

	maxPage := count / pageCount
	if count == pageCount {
		maxPage = 0
	}
	page = utils.Min64(page, maxPage)
	skip := utils.Min64(page*pageCount, count)

	cursor, err := coll.Find(
		db,
		query,
		&options.FindOptions{
			Sort: &bson.D{
				{"name", 1},
			},
			Skip:  &skip,
			Limit: &pageCount,
		},
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		balnc := &Balancer{}
		err = cursor.Decode(balnc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		balncs = append(balncs, balnc)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func Remove(db *database.Database, balncId bson.ObjectID) (err error) {
	coll := db.Balancers()

	_, err = coll.DeleteOne(db, &bson.M{
		"_id": balncId,
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

func RemoveOrg(db *database.Database, orgId, balncId bson.ObjectID) (
	err error) {

	coll := db.Balancers()

	_, err = coll.DeleteOne(db, &bson.M{
		"_id":          balncId,
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

func RemoveMulti(db *database.Database, balncIds []bson.ObjectID) (
	err error) {
	coll := db.Balancers()

	_, err = coll.DeleteMany(db, &bson.M{
		"_id": &bson.M{
			"$in": balncIds,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func RemoveMultiOrg(db *database.Database, orgId bson.ObjectID,
	balncIds []bson.ObjectID) (err error) {

	coll := db.Balancers()

	_, err = coll.DeleteMany(db, &bson.M{
		"_id": &bson.M{
			"$in": balncIds,
		},
		"organization": orgId,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

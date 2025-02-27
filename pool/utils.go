package pool

import (
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/mongo-go-driver/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/utils"
)

func Get(db *database.Database, poolId primitive.ObjectID) (
	pl *Pool, err error) {

	coll := db.Pools()
	pl = &Pool{}

	err = coll.FindOneId(poolId, pl)
	if err != nil {
		return
	}

	return
}

func GetOne(db *database.Database, query *bson.M) (pl *Pool, err error) {
	coll := db.Pools()
	pl = &Pool{}

	err = coll.FindOne(db, query).Decode(pl)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAll(db *database.Database, query *bson.M) (
	pools []*Pool, err error) {

	coll := db.Pools()
	pools = []*Pool{}

	cursor, err := coll.Find(db, query)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		nde := &Pool{}
		err = cursor.Decode(nde)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		pools = append(pools, nde)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllPaged(db *database.Database, query *bson.M,
	page, pageCount int64) (pools []*Pool, count int64, err error) {

	coll := db.Pools()
	pools = []*Pool{}

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
		pl := &Pool{}
		err = cursor.Decode(pl)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		pools = append(pools, pl)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllNames(db *database.Database, query *bson.M) (
	pools []*Pool, err error) {

	coll := db.Pools()
	pools = []*Pool{}

	cursor, err := coll.Find(
		db,
		query,
		&options.FindOptions{
			Sort: &bson.D{
				{"name", 1},
			},
			Projection: &bson.D{
				{"_id", 1},
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
		pl := &Pool{}
		err = cursor.Decode(pl)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		pools = append(pools, pl)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func Remove(db *database.Database, poolId primitive.ObjectID) (err error) {
	coll := db.Pools()

	_, err = coll.DeleteOne(db, &bson.M{
		"_id":               poolId,
		"delete_protection": false,
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

func RemoveMulti(db *database.Database, poolIds []primitive.ObjectID) (
	err error) {

	coll := db.Pools()

	_, err = coll.DeleteMany(db, &bson.M{
		"_id": &bson.M{
			"$in": poolIds,
		},
		"delete_protection": false,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

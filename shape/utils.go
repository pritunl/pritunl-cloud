package shape

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/mongo-go-driver/v2/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/utils"
)

func Get(db *database.Database, shapeId bson.ObjectID) (
	shpe *Shape, err error) {

	coll := db.Shapes()
	shpe = &Shape{}

	err = coll.FindOneId(shapeId, shpe)
	if err != nil {
		return
	}

	return
}

func GetOne(db *database.Database, query *bson.M) (shpe *Shape, err error) {
	coll := db.Shapes()
	shpe = &Shape{}

	err = coll.FindOne(db, query).Decode(shpe)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAll(db *database.Database, query *bson.M) (
	shapes []*Shape, err error) {

	coll := db.Shapes()
	shapes = []*Shape{}

	cursor, err := coll.Find(db, query)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		nde := &Shape{}
		err = cursor.Decode(nde)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		shapes = append(shapes, nde)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllPaged(db *database.Database, query *bson.M,
	page, pageCount int64) (shapes []*Shape, count int64, err error) {

	coll := db.Shapes()
	shapes = []*Shape{}

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
		shpe := &Shape{}
		err = cursor.Decode(shpe)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		shapes = append(shapes, shpe)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllNames(db *database.Database, query *bson.M) (
	shapes []*Shape, err error) {

	coll := db.Shapes()
	shapes = []*Shape{}

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
				{"type", 1},
				{"zone", 1},
				{"flexible", 1},
				{"disk_type", 1},
				{"disk_pool", 1},
				{"memory", 1},
				{"processors", 1},
			},
		},
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		shpe := &Shape{}
		err = cursor.Decode(shpe)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		shapes = append(shapes, shpe)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func Remove(db *database.Database, shapeId bson.ObjectID) (err error) {
	coll := db.Shapes()

	_, err = coll.DeleteOne(db, &bson.M{
		"_id":               shapeId,
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

func RemoveMulti(db *database.Database, shapeIds []bson.ObjectID) (
	err error) {

	coll := db.Shapes()

	_, err = coll.DeleteMany(db, &bson.M{
		"_id": &bson.M{
			"$in": shapeIds,
		},
		"delete_protection": false,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

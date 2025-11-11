package zone

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/mongo-go-driver/v2/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/utils"
)

func Get(db *database.Database, zoneId bson.ObjectID) (
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
		options.Find().
			SetSort(&bson.D{
				{"name", 1},
			}),
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

func GetAllNames(db *database.Database, query *bson.M) (
	znes []*Completion, err error) {

	coll := db.Certificates()
	znes = []*Completion{}

	cursor, err := coll.Find(
		db,
		query,
		options.Find().
			SetSort(&bson.D{
				{"name", 1},
			}).
			SetProjection(&bson.D{
				{"name", 1},
				{"datacenter", 1},
			}),
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		zne := &Completion{}
		err = cursor.Decode(zne)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		znes = append(znes, zne)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllPaged(db *database.Database, query *bson.M,
	page, pageCount int64) (znes []*Zone, count int64, err error) {

	coll := db.Zones()
	znes = []*Zone{}

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
			SetSort(&bson.D{
				{"name", 1},
			}).
			SetSkip(skip).
			SetLimit(pageCount),
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

		znes = append(znes, zne)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllDatacenter(db *database.Database, dcId bson.ObjectID) (
	zones []*Zone, err error) {

	coll := db.Zones()
	zones = []*Zone{}

	cursor, err := coll.Find(
		db,
		&bson.M{
			"datacenter": dcId,
		},
		options.Find().
			SetSort(&bson.D{
				{"name", 1},
			}),
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

func GetAllNamedDc(db *database.Database, dcIds []bson.ObjectID) (
	zones []*Completion, err error) {

	coll := db.Zones()
	zones = []*Completion{}

	cursor, err := coll.Find(
		db,
		&bson.M{
			"datacenter": &bson.M{
				"$in": dcIds,
			},
		},
		options.Find().
			SetSort(&bson.D{
				{"name", 1},
			}).
			SetProjection(&bson.D{
				{"name", 1},
				{"datacenter", 1},
			}),
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		zne := &Completion{}
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

func Remove(db *database.Database, zoneId bson.ObjectID) (err error) {
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

func RemoveMulti(db *database.Database, zneIds []bson.ObjectID) (
	err error) {
	coll := db.Zones()

	_, err = coll.DeleteMany(db, &bson.M{
		"_id": &bson.M{
			"$in": zneIds,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

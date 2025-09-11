package datacenter

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/mongo-go-driver/v2/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/utils"
)

func Get(db *database.Database, dcId bson.ObjectID) (
	dc *Datacenter, err error) {

	coll := db.Datacenters()
	dc = &Datacenter{}

	err = coll.FindOneId(dcId, dc)
	if err != nil {
		return
	}

	return
}

func ExistsOrg(db *database.Database, orgId, dcId bson.ObjectID) (
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

func GetAllNamesOrg(db *database.Database, orgId bson.ObjectID) (
	dcs []*Completion, err error) {

	coll := db.Datacenters()
	dcs = []*Completion{}

	cursor, err := coll.Find(db, &bson.M{
		"$or": []*bson.M{
			&bson.M{
				"match_organizations": false,
			},
			&bson.M{
				"organizations": orgId,
			},
		},
	}, &options.FindOptions{
		Sort: &bson.D{
			{"name", 1},
		},
		Projection: &bson.D{
			{"name", 1},
			{"network_mode", 1},
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		dc := &Completion{}
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

func GetAllNames(db *database.Database, query *bson.M) (
	dcs []*Completion, err error) {

	coll := db.Certificates()
	dcs = []*Completion{}

	cursor, err := coll.Find(
		db,
		query,
		&options.FindOptions{
			Sort: &bson.D{
				{"name", 1},
			},
			Projection: &bson.D{
				{"name", 1},
				{"network_mode", 1},
			},
		},
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		dc := &Completion{}
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

func GetAllPaged(db *database.Database, query *bson.M,
	page, pageCount int64) (dc []*Datacenter, count int64, err error) {

	coll := db.Datacenters()
	dc = []*Datacenter{}

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
		d := &Datacenter{}
		err = cursor.Decode(d)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		dc = append(dc, d)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func DistinctOrg(db *database.Database, orgId bson.ObjectID) (
	ids []bson.ObjectID, err error) {

	coll := db.Datacenters()
	ids = []bson.ObjectID{}

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
		if id, ok := idInf.(bson.ObjectID); ok {
			ids = append(ids, id)
		}
	}

	return
}

func Remove(db *database.Database, dcId bson.ObjectID) (err error) {
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

func RemoveMulti(db *database.Database, dcIds []bson.ObjectID) (
	err error) {
	coll := db.Datacenters()

	_, err = coll.DeleteMany(db, &bson.M{
		"_id": &bson.M{
			"$in": dcIds,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func RemoveMultiOrg(db *database.Database, orgId bson.ObjectID,
	dcIds []bson.ObjectID) (err error) {

	coll := db.Datacenters()

	_, err = coll.DeleteMany(db, &bson.M{
		"_id": &bson.M{
			"$in": dcIds,
		},
		"organization": orgId,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

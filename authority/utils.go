package authority

import (
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/mongo-go-driver/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/utils"
)

func Get(db *database.Database, authrId primitive.ObjectID) (
	authr *Authority, err error) {

	coll := db.Authorities()
	authr = &Authority{}

	err = coll.FindOneId(authrId, authr)
	if err != nil {
		return
	}

	return
}

func GetOrg(db *database.Database, orgId, authrId primitive.ObjectID) (
	authr *Authority, err error) {

	coll := db.Authorities()
	authr = &Authority{}

	err = coll.FindOne(db, &bson.M{
		"_id":          authrId,
		"organization": orgId,
	}).Decode(authr)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAll(db *database.Database, query *bson.M) (
	authrs []*Authority, err error) {

	coll := db.Authorities()
	authrs = []*Authority{}

	cursor, err := coll.Find(db, query)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		authr := &Authority{}
		err = cursor.Decode(authr)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		authrs = append(authrs, authr)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetRoles(db *database.Database, roles []string) (
	authrs []*Authority, err error) {

	coll := db.Authorities()
	authrs = []*Authority{}

	cursor, err := coll.Find(db, &bson.M{
		"organization": Global,
		"network_roles": &bson.M{
			"$in": roles,
		},
	}, &options.FindOptions{
		Sort: &bson.D{
			{"_id", 1},
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		authr := &Authority{}
		err = cursor.Decode(authr)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		authrs = append(authrs, authr)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetMapRoles(db *database.Database, query *bson.M) (
	authrs map[string][]*Authority, err error) {

	coll := db.Authorities()
	authrs = map[string][]*Authority{}

	cursor, err := coll.Find(db, query)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		authr := &Authority{}
		err = cursor.Decode(authr)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		for _, role := range authr.NetworkRoles {
			roleAuthrs := authrs[role]
			if roleAuthrs == nil {
				roleAuthrs = []*Authority{}
			}
			authrs[role] = append(roleAuthrs, authr)
		}
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetOrgMapRoles(db *database.Database, orgId primitive.ObjectID) (
	authrs map[string][]*Authority, err error) {

	coll := db.Authorities()
	authrs = map[string][]*Authority{}

	cursor, err := coll.Find(db, &bson.M{
		"organization": orgId,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		authr := &Authority{}
		err = cursor.Decode(authr)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		for _, role := range authr.NetworkRoles {
			roleAuthrs := authrs[role]
			if roleAuthrs == nil {
				roleAuthrs = []*Authority{}
			}
			authrs[role] = append(roleAuthrs, authr)
		}
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetOrgRoles(db *database.Database, orgId primitive.ObjectID,
	roles []string) (authrs []*Authority, err error) {

	coll := db.Authorities()
	authrs = []*Authority{}

	cursor, err := coll.Find(db, &bson.M{
		"organization": orgId,
		"network_roles": &bson.M{
			"$in": roles,
		},
	}, &options.FindOptions{
		Sort: &bson.D{
			{"_id", 1},
		},
	})
	defer cursor.Close(db)

	for cursor.Next(db) {
		authr := &Authority{}
		err = cursor.Decode(authr)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		authrs = append(authrs, authr)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllPaged(db *database.Database, query *bson.M,
	page, pageCount int64) (authrs []*Authority, count int64, err error) {

	coll := db.Authorities()
	authrs = []*Authority{}

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
		authr := &Authority{}
		err = cursor.Decode(authr)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		authrs = append(authrs, authr)
		authr = &Authority{}
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func Remove(db *database.Database, authrId primitive.ObjectID) (err error) {
	coll := db.Authorities()

	_, err = coll.DeleteOne(db, &bson.M{
		"_id": authrId,
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

func RemoveOrg(db *database.Database, orgId, authrId primitive.ObjectID) (
	err error) {

	coll := db.Authorities()

	_, err = coll.DeleteOne(db, &bson.M{
		"_id":          authrId,
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

func RemoveMulti(db *database.Database,
	authrIds []primitive.ObjectID) (err error) {

	coll := db.Authorities()

	_, err = coll.DeleteMany(db, &bson.M{
		"_id": &bson.M{
			"$in": authrIds,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func RemoveMultiOrg(db *database.Database, orgId primitive.ObjectID,
	authrIds []primitive.ObjectID) (err error) {

	coll := db.Authorities()

	_, err = coll.DeleteMany(db, &bson.M{
		"_id": &bson.M{
			"$in": authrIds,
		},
		"organization": orgId,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

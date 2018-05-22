package authority

import (
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/utils"
	"gopkg.in/mgo.v2/bson"
)

func Get(db *database.Database, authrId bson.ObjectId) (
	authr *Authority, err error) {

	coll := db.Authorities()
	authr = &Authority{}

	err = coll.FindOneId(authrId, authr)
	if err != nil {
		return
	}

	return
}

func GetOrg(db *database.Database, orgId, authrId bson.ObjectId) (
	authr *Authority, err error) {

	coll := db.Authorities()
	authr = &Authority{}

	err = coll.FindOne(&bson.M{
		"_id":          authrId,
		"organization": orgId,
	}, authr)
	if err != nil {
		return
	}

	return
}

func GetAll(db *database.Database, query *bson.M) (
	authrs []*Authority, err error) {

	coll := db.Authorities()
	authrs = []*Authority{}

	cursor := coll.Find(query).Iter()

	nde := &Authority{}
	for cursor.Next(nde) {
		authrs = append(authrs, nde)
		nde = &Authority{}
	}

	err = cursor.Close()
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

	cursor := coll.Find(&bson.M{
		"$or": []*bson.M{
			&bson.M{
				"organization": nil,
			},
			&bson.M{
				"organization": &bson.M{
					"$exists": false,
				},
			},
		},
		"network_roles": &bson.M{
			"$in": roles,
		},
	}).Sort("_id").Iter()

	nde := &Authority{}
	for cursor.Next(nde) {
		authrs = append(authrs, nde)
		nde = &Authority{}
	}

	err = cursor.Close()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetOrgMapRoles(db *database.Database, orgId bson.ObjectId) (
	authrs map[string][]*Authority, err error) {

	coll := db.Authorities()
	authrs = map[string][]*Authority{}

	cursor := coll.Find(&bson.M{
		"organization": orgId,
	}).Iter()

	authr := &Authority{}
	for cursor.Next(authr) {
		for _, role := range authr.NetworkRoles {
			roleAuthrs := authrs[role]
			if roleAuthrs == nil {
				roleAuthrs = []*Authority{}
			}
			authrs[role] = append(roleAuthrs, authr)
		}
		authr = &Authority{}
	}

	err = cursor.Close()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetOrgRoles(db *database.Database, orgId bson.ObjectId,
	roles []string) (authrs []*Authority, err error) {

	coll := db.Authorities()
	authrs = []*Authority{}

	cursor := coll.Find(&bson.M{
		"organization": orgId,
		"network_roles": &bson.M{
			"$in": roles,
		},
	}).Sort("_id").Iter()

	nde := &Authority{}
	for cursor.Next(nde) {
		authrs = append(authrs, nde)
		nde = &Authority{}
	}

	err = cursor.Close()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllPaged(db *database.Database, query *bson.M, page, pageCount int) (
	authrs []*Authority, count int, err error) {

	coll := db.Authorities()
	authrs = []*Authority{}

	qury := coll.Find(query)

	count, err = qury.Count()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	skip := utils.Min(page*pageCount, utils.Max(0, count-pageCount))

	cursor := qury.Sort("name").Skip(skip).Limit(pageCount).Iter()

	authr := &Authority{}
	for cursor.Next(authr) {
		authrs = append(authrs, authr)
		authr = &Authority{}
	}

	err = cursor.Close()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func Remove(db *database.Database, authrId bson.ObjectId) (err error) {
	coll := db.Authorities()

	err = coll.Remove(&bson.M{
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

func RemoveOrg(db *database.Database, orgId, authrId bson.ObjectId) (err error) {
	coll := db.Authorities()

	err = coll.Remove(&bson.M{
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
	authrIds []bson.ObjectId) (err error) {

	coll := db.Authorities()

	_, err = coll.RemoveAll(&bson.M{
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

func RemoveMultiOrg(db *database.Database, orgId bson.ObjectId,
	authrIds []bson.ObjectId) (err error) {

	coll := db.Authorities()

	_, err = coll.RemoveAll(&bson.M{
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

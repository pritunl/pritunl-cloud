package domain

import (
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/utils"
	"gopkg.in/mgo.v2/bson"
)

func Get(db *database.Database, domnId bson.ObjectId) (
	domn *Domain, err error) {

	coll := db.Domains()
	domn = &Domain{}

	err = coll.FindOneId(domnId, domn)
	if err != nil {
		return
	}

	return
}

func GetOrg(db *database.Database, orgId, domnId bson.ObjectId) (
	domn *Domain, err error) {

	coll := db.Domains()
	domn = &Domain{}

	err = coll.FindOne(&bson.M{
		"_id":          domnId,
		"organization": orgId,
	}, domn)
	if err != nil {
		return
	}

	return
}

func ExistsOrg(db *database.Database, orgId, domnId bson.ObjectId) (
	exists bool, err error) {

	coll := db.Domains()

	n, err := coll.Find(&bson.M{
		"_id":          domnId,
		"organization": orgId,
	}).Count()
	if err != nil {
		return
	}

	if n > 0 {
		exists = true
	}

	return
}

func GetAll(db *database.Database, query *bson.M) (
	domns []*Domain, err error) {

	coll := db.Domains()
	domns = []*Domain{}

	cursor := coll.Find(query).Iter()

	nde := &Domain{}
	for cursor.Next(nde) {
		domns = append(domns, nde)
		nde = &Domain{}
	}

	err = cursor.Close()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllPaged(db *database.Database, query *bson.M, page, pageCount int) (
	domns []*Domain, count int, err error) {

	coll := db.Domains()
	domns = []*Domain{}

	qury := coll.Find(query)

	count, err = qury.Count()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	skip := utils.Min(page*pageCount, utils.Max(0, count-pageCount))

	cursor := qury.Sort("name").Skip(skip).Limit(pageCount).Iter()

	domn := &Domain{}
	for cursor.Next(domn) {
		domns = append(domns, domn)
		domn = &Domain{}
	}

	err = cursor.Close()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func Remove(db *database.Database, domnId bson.ObjectId) (err error) {
	coll := db.Domains()

	err = coll.Remove(&bson.M{
		"_id": domnId,
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

func RemoveOrg(db *database.Database, orgId, domnId bson.ObjectId) (
	err error) {

	coll := db.Domains()

	err = coll.Remove(&bson.M{
		"_id":          domnId,
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

func RemoveMulti(db *database.Database, domnIds []bson.ObjectId) (err error) {
	coll := db.Domains()

	_, err = coll.RemoveAll(&bson.M{
		"_id": &bson.M{
			"$in": domnIds,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func RemoveMultiOrg(db *database.Database, orgId bson.ObjectId,
	domnIds []bson.ObjectId) (err error) {

	coll := db.Domains()

	_, err = coll.RemoveAll(&bson.M{
		"_id": &bson.M{
			"$in": domnIds,
		},
		"organization": orgId,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetRecordAll(db *database.Database, query *bson.M) (
	recrds []*Record, err error) {

	coll := db.DomainsRecord()
	recrds = []*Record{}

	cursor := coll.Find(query).Iter()

	recrd := &Record{}
	for cursor.Next(recrd) {
		recrds = append(recrds, recrd)
		recrd = &Record{}
	}

	err = cursor.Close()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func RemoveRecord(db *database.Database, recrdId bson.ObjectId) (err error) {
	coll := db.DomainsRecord()

	err = coll.Remove(&bson.M{
		"_id": recrdId,
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

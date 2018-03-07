package vpc

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/utils"
	"gopkg.in/mgo.v2/bson"
)

func Get(db *database.Database, vcId bson.ObjectId) (
	vc *Vpc, err error) {

	coll := db.Vpcs()
	vc = &Vpc{}

	err = coll.FindOneId(vcId, vc)
	if err != nil {
		return
	}

	return
}

func GetAll(db *database.Database, query *bson.M) (
	vcs []*Vpc, err error) {

	coll := db.Vpcs()
	vcs = []*Vpc{}

	cursor := coll.Find(query).Iter()

	nde := &Vpc{}
	for cursor.Next(nde) {
		vcs = append(vcs, nde)
		nde = &Vpc{}
	}

	err = cursor.Close()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllNames(db *database.Database, query *bson.M) (
	vpcs []*Vpc, err error) {

	coll := db.Vpcs()
	vpcs = []*Vpc{}

	cursor := coll.Find(query).Sort("name").Select(&bson.M{
		"name":         1,
		"organization": 1,
		"type":         1,
	}).Iter()

	vc := &Vpc{}
	for cursor.Next(vc) {
		vpcs = append(vpcs, vc)
		vc = &Vpc{}
	}

	err = cursor.Close()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllPaged(db *database.Database, query *bson.M, page, pageCount int) (
	vcs []*Vpc, count int, err error) {

	coll := db.Vpcs()
	vcs = []*Vpc{}

	qury := coll.Find(query)

	count, err = qury.Count()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	skip := utils.Min(page*pageCount, utils.Max(0, count-pageCount))

	cursor := qury.Sort("name").Skip(skip).Limit(pageCount).Iter()

	vc := &Vpc{}
	for cursor.Next(vc) {
		vcs = append(vcs, vc)
		vc = &Vpc{}
	}

	err = cursor.Close()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func DistinctIds(db *database.Database, matchIds []bson.ObjectId) (
	idsSet set.Set, err error) {

	coll := db.Images()

	idsSet = set.NewSet()

	ids := []bson.ObjectId{}
	err = coll.Find(&bson.M{
		"_id": &bson.M{
			"$in": matchIds,
		},
	}).Distinct("_id", &ids)
	if err != nil {
		return
	}

	for _, id := range ids {
		idsSet.Add(id)
	}

	return
}

func Remove(db *database.Database, vcId bson.ObjectId) (err error) {
	coll := db.Vpcs()

	err = coll.Remove(&bson.M{
		"_id": vcId,
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

func RemoveMulti(db *database.Database, vcIds []bson.ObjectId) (err error) {
	coll := db.Vpcs()

	_, err = coll.RemoveAll(&bson.M{
		"_id": &bson.M{
			"$in": vcIds,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

package instance

import (
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/utils"
	"gopkg.in/mgo.v2/bson"
)

func Get(db *database.Database, instId bson.ObjectId) (
	inst *Instance, err error) {

	coll := db.Instances()
	inst = &Instance{}

	err = coll.FindOneId(instId, inst)
	if err != nil {
		return
	}

	return
}

func GetAll(db *database.Database) (insts []*Instance, err error) {
	coll := db.Instances()
	insts = []*Instance{}

	cursor := coll.Find(bson.M{}).Iter()

	nde := &Instance{}
	for cursor.Next(nde) {
		insts = append(insts, nde)
		nde = &Instance{}
	}

	err = cursor.Close()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllPaged(db *database.Database, query *bson.M, page, pageCount int) (
	insts []*Instance, count int, err error) {

	coll := db.Instances()
	insts = []*Instance{}

	qury := coll.Find(query)

	count, err = qury.Count()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	skip := utils.Min(page*pageCount, utils.Max(0, count-pageCount))

	cursor := qury.Sort("name").Skip(skip).Limit(pageCount).Iter()

	inst := &Instance{}
	for cursor.Next(inst) {
		insts = append(insts, inst)
		inst = &Instance{}
	}

	err = cursor.Close()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func Remove(db *database.Database, instId bson.ObjectId) (err error) {
	coll := db.Instances()

	err = coll.Remove(&bson.M{
		"_id": instId,
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

func RemoveMulti(db *database.Database, instIds []bson.ObjectId) (err error) {
	coll := db.Instances()

	_, err = coll.RemoveAll(&bson.M{
		"_id": &bson.M{
			"$in": instIds,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

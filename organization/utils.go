package organization

import (
	"github.com/pritunl/pritunl-cloud/database"
	"gopkg.in/mgo.v2/bson"
)

func Get(db *database.Database, dcId bson.ObjectId) (
	dc *Organization, err error) {

	coll := db.Organizations()
	dc = &Organization{}

	err = coll.FindOneId(dcId, dc)
	if err != nil {
		return
	}

	return
}

func GetAll(db *database.Database) (dcs []*Organization, err error) {
	coll := db.Organizations()
	dcs = []*Organization{}

	cursor := coll.Find(bson.M{}).Iter()

	nde := &Organization{}
	for cursor.Next(nde) {
		dcs = append(dcs, nde)
		nde = &Organization{}
	}

	err = cursor.Close()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func Remove(db *database.Database, dcId bson.ObjectId) (err error) {
	coll := db.Organizations()

	err = coll.Remove(&bson.M{
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

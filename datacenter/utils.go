package datacenter

import (
	"github.com/pritunl/pritunl-cloud/database"
	"gopkg.in/mgo.v2/bson"
)

func Get(db *database.Database, dcId bson.ObjectId) (
	dc *Datacenter, err error) {

	coll := db.Datacenters()
	dc = &Datacenter{}

	err = coll.FindOneId(dcId, dc)
	if err != nil {
		return
	}

	return
}

func ExistsOrg(db *database.Database, orgId, dcId bson.ObjectId) (
	exists bool, err error) {

	coll := db.Datacenters()

	n, err := coll.Find(&bson.M{
		"_id": dcId,
		"$or": []*bson.M{
			&bson.M{
				"match_organizations": false,
			},
			&bson.M{
				"organizations": orgId,
			},
		},
	}).Count()
	if err != nil {
		return
	}

	if n > 0 {
		exists = true
	}

	return
}

func GetAll(db *database.Database) (dcs []*Datacenter, err error) {
	coll := db.Datacenters()
	dcs = []*Datacenter{}

	cursor := coll.Find(bson.M{}).Iter()

	nde := &Datacenter{}
	for cursor.Next(nde) {
		dcs = append(dcs, nde)
		nde = &Datacenter{}
	}

	err = cursor.Close()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllNamesOrg(db *database.Database, orgId bson.ObjectId) (
	dcs []*Datacenter, err error) {

	coll := db.Datacenters()
	dcs = []*Datacenter{}

	cursor := coll.Find(bson.M{
		"$or": []*bson.M{
			&bson.M{
				"match_organizations": false,
			},
			&bson.M{
				"organizations": orgId,
			},
		},
	}).Iter()

	nde := &Datacenter{}
	for cursor.Next(nde) {
		dcs = append(dcs, nde)
		nde = &Datacenter{}
	}

	err = cursor.Close()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func Remove(db *database.Database, dcId bson.ObjectId) (err error) {
	coll := db.Datacenters()

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

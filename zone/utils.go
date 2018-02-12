package zone

import (
	"github.com/pritunl/pritunl-cloud/database"
	"gopkg.in/mgo.v2/bson"
)

func Get(db *database.Database, zoneId bson.ObjectId) (
	zone *Zone, err error) {

	coll := db.Zones()
	zone = &Zone{}

	err = coll.FindOneId(zoneId, zone)
	if err != nil {
		return
	}

	return
}

func GetAll(db *database.Database) (zones []*Zone, err error) {
	coll := db.Zones()
	zones = []*Zone{}

	cursor := coll.Find(bson.M{}).Iter()

	nde := &Zone{}
	for cursor.Next(nde) {
		zones = append(zones, nde)
		nde = &Zone{}
	}

	err = cursor.Close()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func Remove(db *database.Database, zoneId bson.ObjectId) (err error) {
	coll := db.Zones()

	err = coll.Remove(&bson.M{
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

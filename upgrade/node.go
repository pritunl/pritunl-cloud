package upgrade

import (
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/pritunl-cloud/database"
)

func nodeUpgrade(db *database.Database) (err error) {
	coll := db.Nodes()
	_, err = coll.UpdateMany(db, bson.M{
		"available_interfaces": bson.M{
			"$exists": true,
		},
		"available_interfaces.0": bson.M{
			"$type": "string",
		},
	}, []bson.M{
		bson.M{
			"$unset": "available_interfaces",
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	_, err = coll.UpdateMany(db, bson.M{
		"available_bridges": bson.M{
			"$exists": true,
		},
		"available_bridges.0": bson.M{
			"$type": "string",
		},
	}, []bson.M{
		bson.M{
			"$unset": "available_bridges",
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

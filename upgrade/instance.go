package upgrade

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
)

func instanceUpgrade(db *database.Database) (err error) {
	coll := db.Instances()
	_, err = coll.UpdateMany(db, bson.M{
		"virt_timestamp": bson.M{
			"$exists": true,
		},
	}, []bson.M{
		bson.M{
			"$set": bson.M{
				"timestamp": "$virt_timestamp",
			},
		},
		bson.M{
			"$unset": "virt_timestamp",
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

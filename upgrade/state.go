package upgrade

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
)

func instStateUpgrade(db *database.Database) (err error) {
	coll := db.Instances()

	_, err = coll.UpdateMany(db, bson.M{
		"virt_state": bson.M{
			"$exists": true,
		},
	}, []bson.M{
		bson.M{
			"$set": bson.M{
				"state": "$virt_state",
			},
		},
		bson.M{
			"$unset": "virt_state",
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

package upgrade

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
)

func rolesUpgrade(db *database.Database) (err error) {
	coll := db.Instances()
	_, err = coll.UpdateMany(db, bson.M{
		"network_roles": bson.M{
			"$exists": true,
		},
	}, []bson.M{
		bson.M{
			"$set": bson.M{
				"roles": "$network_roles",
			},
		},
		bson.M{
			"$unset": "network_roles",
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	coll = db.Nodes()
	_, err = coll.UpdateMany(db, bson.M{
		"network_roles": bson.M{
			"$exists": true,
		},
	}, []bson.M{
		bson.M{
			"$set": bson.M{
				"roles": "$network_roles",
			},
		},
		bson.M{
			"$unset": "network_roles",
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	coll = db.Firewalls()
	_, err = coll.UpdateMany(db, bson.M{
		"network_roles": bson.M{
			"$exists": true,
		},
	}, []bson.M{
		bson.M{
			"$set": bson.M{
				"roles": "$network_roles",
			},
		},
		bson.M{
			"$unset": "network_roles",
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	coll = db.Authorities()
	_, err = coll.UpdateMany(db, bson.M{
		"network_roles": bson.M{
			"$exists": true,
		},
	}, []bson.M{
		bson.M{
			"$set": bson.M{
				"roles": "$network_roles",
			},
		},
		bson.M{
			"$unset": "network_roles",
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

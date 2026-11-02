package upgrade

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
)

func runtimeMerge(db *database.Database, coll *database.Collection,
	match *bson.M, root *bson.M, into string) (err error) {

	cursor, err := coll.Aggregate(db, []*bson.M{
		&bson.M{
			"$match": match,
		},
		&bson.M{
			"$replaceRoot": &bson.M{
				"newRoot": root,
			},
		},
		&bson.M{
			"$merge": &bson.M{
				"into":           into,
				"on":             "_id",
				"whenMatched":    "merge",
				"whenNotMatched": "insert",
			},
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func runtimeUpgrade(db *database.Database) (err error) {
	coll := db.Nodes()

	nodeMatch := &bson.M{
		"$or": []*bson.M{
			&bson.M{
				"metric": &bson.M{
					"$exists": true,
				},
			},
			&bson.M{
				"timestamp": &bson.M{
					"$exists": true,
				},
			},
		},
	}

	err = runtimeMerge(db, coll, nodeMatch, &bson.M{
		"$mergeObjects": []interface{}{
			&bson.M{
				"_id":       "$_id",
				"timestamp": "$timestamp",
			},
			"$metric",
		},
	}, "nodes_runtime")
	if err != nil {
		return
	}

	_, err = coll.UpdateMany(db, nodeMatch, &bson.M{
		"$unset": &bson.M{
			"metric":    "",
			"timestamp": "",
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	coll = db.Instances()

	err = runtimeMerge(db, coll, &bson.M{
		"guest": &bson.M{
			"$exists": true,
		},
	}, &bson.M{
		"$mergeObjects": []interface{}{
			&bson.M{
				"_id": "$_id",
			},
			"$guest",
		},
	}, "instances_runtime")
	if err != nil {
		return
	}

	err = runtimeMerge(db, coll, &bson.M{
		"info": &bson.M{
			"$exists": true,
		},
	}, &bson.M{
		"$mergeObjects": []interface{}{
			&bson.M{
				"_id": "$_id",
			},
			"$info",
		},
	}, "instances_info")
	if err != nil {
		return
	}

	_, err = coll.UpdateMany(db, &bson.M{
		"$or": []*bson.M{
			&bson.M{
				"guest": &bson.M{
					"$exists": true,
				},
			},
			&bson.M{
				"info": &bson.M{
					"$exists": true,
				},
			},
		},
	}, &bson.M{
		"$unset": &bson.M{
			"guest": "",
			"info":  "",
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

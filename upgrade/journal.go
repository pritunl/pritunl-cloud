package upgrade

import (
	"time"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/mongo-go-driver/v2/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/journal"
)

func journalUpgrade(db *database.Database) (err error) {
	coll := db.Journal()

	cursor, err := coll.Find(db, &bson.M{
		"c": &bson.M{
			"$exists": false,
		},
	}, options.Find().
		SetSort(bson.D{
			{"t", 1},
			{"_id", 1},
		}),
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	var count int32
	var lastTime time.Time
	i := 0

	for cursor.Next(db) {
		jrnl := &journal.Journal{}
		err = cursor.Decode(jrnl)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		if jrnl.Timestamp.Unix() != lastTime.Unix() {
			count = 1
		}
		lastTime = jrnl.Timestamp

		if i%1000 == 0 {
			println(count)
		}
		i += 1

		_, err = coll.UpdateOne(db, &bson.M{
			"_id": jrnl.Id,
		}, &bson.M{
			"$set": &bson.M{
				"c": count,
			},
		})
		if err != nil {
			err = database.ParseError(err)
			return
		}

		count += 1
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

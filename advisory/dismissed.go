package advisory

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/mongo-go-driver/v2/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
)

type Dismissal struct {
	Dismissed          bool
	DismissedResources []bson.ObjectID
}

func GetDismissals(db *database.Database) (
	dismissals map[bson.ObjectID]map[string]*Dismissal, err error) {

	coll := db.Advisories()
	dismissals = map[bson.ObjectID]map[string]*Dismissal{}

	cursor, err := coll.Find(
		db,
		&bson.M{},
		options.Find().SetProjection(&bson.M{
			"organization":        1,
			"reference":           1,
			"dismissed":           1,
			"dismissed_resources": 1,
		}),
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		adv := &Advisory{}
		err = cursor.Decode(adv)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		orgDismissals := dismissals[adv.Organization]
		if orgDismissals == nil {
			orgDismissals = map[string]*Dismissal{}
			dismissals[adv.Organization] = orgDismissals
		}

		orgDismissals[adv.Reference] = &Dismissal{
			Dismissed:          adv.Dismissed,
			DismissedResources: adv.DismissedResources,
		}
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

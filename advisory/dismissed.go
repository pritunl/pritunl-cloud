package advisory

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/mongo-go-driver/v2/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
)

type Dismissal struct {
	Dismissed bool
	Resources set.Set
}

func (d *Dismissal) HasResource(resourceId bson.ObjectID) bool {
	if d == nil || d.Resources == nil {
		return false
	}
	return d.Resources.Contains(resourceId)
}

func GetDismissals(db *database.Database) (
	dismissals map[bson.ObjectID]map[string]*Dismissal, err error) {

	coll := db.Advisories()
	dismissals = map[bson.ObjectID]map[string]*Dismissal{}

	getDismissal := func(orgId bson.ObjectID, ref string) *Dismissal {
		orgDismissals := dismissals[orgId]
		if orgDismissals == nil {
			orgDismissals = map[string]*Dismissal{}
			dismissals[orgId] = orgDismissals
		}

		dism := orgDismissals[ref]
		if dism == nil {
			dism = &Dismissal{
				Resources: set.NewSet(),
			}
			orgDismissals[ref] = dism
		}

		return dism
	}

	cursor, err := coll.Find(
		db,
		&bson.M{
			"dismissed": true,
		},
		options.Find().SetProjection(&bson.M{
			"organization": 1,
			"reference":    1,
			"dismissed":    1,
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

		getDismissal(adv.Organization, adv.Reference).Dismissed = true
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	resources, err := GetResources(db, &bson.M{
		"dismissed": true,
	}, options.Find().SetProjection(&bson.M{
		"organization": 1,
		"reference":    1,
		"resource":     1,
	}))
	if err != nil {
		return
	}

	for _, res := range resources {
		getDismissal(res.Organization, res.Reference).Resources.Add(
			res.Resource)
	}

	return
}

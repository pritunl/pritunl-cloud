package advisory

import (
	"time"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/mongo-go-driver/v2/mongo"
	"github.com/pritunl/mongo-go-driver/v2/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
)

type Resource struct {
	Id           bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Organization bson.ObjectID `bson:"organization" json:"organization"`
	Reference    string        `bson:"reference" json:"reference"`
	Resource     bson.ObjectID `bson:"resource" json:"resource"`
	Kind         string        `bson:"kind" json:"kind"`
	State        string        `bson:"state" json:"state"`
	Dismissed    bool          `bson:"dismissed" json:"dismissed"`
	Exclusions   []string      `bson:"exclusions" json:"exclusions"`
	Timestamp    time.Time     `bson:"timestamp" json:"timestamp"`
}

func (r *Resource) UpsertModel() mongo.WriteModel {
	exclusions := r.Exclusions
	if exclusions == nil {
		exclusions = []string{}
	}

	return mongo.NewUpdateOneModel().
		SetFilter(&bson.M{
			"reference": r.Reference,
			"resource":  r.Resource,
		}).
		SetUpdate(&bson.M{
			"$set": &bson.M{
				"organization": r.Organization,
				"reference":    r.Reference,
				"resource":     r.Resource,
				"kind":         r.Kind,
				"state":        r.State,
				"exclusions":   exclusions,
				"timestamp":    r.Timestamp,
			},
			"$setOnInsert": &bson.M{
				"dismissed": false,
			},
		}).
		SetUpsert(true)
}

func UpsertResources(db *database.Database,
	resources []*Resource) (err error) {

	if len(resources) == 0 {
		return
	}

	models := make([]mongo.WriteModel, 0, len(resources))
	for _, res := range resources {
		models = append(models, res.UpsertModel())
	}

	coll := db.AdvisoryResources()

	_, err = coll.BulkWrite(db, models,
		options.BulkWrite().SetOrdered(false))
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetResources(db *database.Database, query *bson.M,
	opts ...options.Lister[options.FindOptions]) (
	resources []*Resource, err error) {

	coll := db.AdvisoryResources()
	resources = []*Resource{}

	cursor, err := coll.Find(db, query, opts...)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		res := &Resource{}
		err = cursor.Decode(res)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		resources = append(resources, res)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func SweepResources(db *database.Database, now time.Time,
	exists func(orgId bson.ObjectID, reference string) bool) (err error) {

	coll := db.AdvisoryResources()

	_, err = coll.DeleteMany(db, &bson.M{
		"timestamp": &bson.M{
			"$lt": now,
		},
		"dismissed": &bson.M{
			"$ne": true,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	dismissed, err := GetResources(db, &bson.M{
		"timestamp": &bson.M{
			"$lt": now,
		},
		"dismissed": true,
	}, options.Find().SetProjection(&bson.M{
		"organization": 1,
		"reference":    1,
	}))
	if err != nil {
		return
	}

	removeIds := []bson.ObjectID{}
	for _, res := range dismissed {
		if !exists(res.Organization, res.Reference) {
			removeIds = append(removeIds, res.Id)
		}
	}

	if len(removeIds) == 0 {
		return
	}

	_, err = coll.DeleteMany(db, &bson.M{
		"_id": &bson.M{
			"$in": removeIds,
		},
		"timestamp": &bson.M{
			"$lt": now,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

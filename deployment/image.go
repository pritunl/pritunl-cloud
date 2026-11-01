package deployment

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
)

func SetImageReady(db *database.Database, deplyId bson.ObjectID) (
	updated bool, err error) {

	coll := db.Deployments()

	resp, err := coll.UpdateOne(db, &bson.M{
		"_id":  deplyId,
		"kind": Image,
		"image_data.state": &bson.M{
			"$in": []interface{}{nil, ""},
		},
	}, &bson.M{
		"$set": &bson.M{
			"image_data.state": Ready,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	updated = resp.ModifiedCount > 0

	return
}

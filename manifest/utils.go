package manifest

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
)

func RemoveResource(db *database.Database,
	resource bson.ObjectID) (err error) {

	coll := db.Manifests()

	_, err = coll.DeleteMany(db, &bson.M{
		"resource": resource,
	})
	if err != nil {
		err = database.ParseError(err)
		switch err.(type) {
		case *database.NotFoundError:
			err = nil
		default:
			return
		}
	}

	return
}

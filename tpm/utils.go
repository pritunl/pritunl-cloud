package tpm

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/mongo-go-driver/v2/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/utils"
)

type instanceData struct {
	Id        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	TpmSecret string        `bson:"tpm_secret" json:"-"`
}

func GenerateSecret() (secret string, err error) {
	secret, err = utils.RandPasswd(128)
	if err != nil {
		return
	}

	return
}

func GetSecret(db *database.Database, vmId bson.ObjectID) (
	secret string, err error) {

	coll := db.Instances()

	data := &instanceData{}

	err = coll.FindOne(
		db,
		&bson.M{
			"_id": vmId,
		},
		&options.FindOneOptions{
			Projection: &bson.D{
				{"tpm_secret", 1},
			},
		},
	).Decode(data)

	secret = data.TpmSecret

	return
}

package imds

import (
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/imds/state"
	"github.com/pritunl/pritunl-cloud/instance"
)

func Sync(db *database.Database, inst *instance.Instance) (err error) {
	data, err := state.Read(inst.Id)
	if err != nil {
		return
	}

	if data == nil {
		return
	}

	coll := db.Instances()

	_, err = coll.UpdateOne(db, &bson.M{
		"_id": inst.Id,
	}, bson.M{
		"$set": &bson.M{
			"guest": &instance.GuestData{
				Heartbeat: data.Timestamp,
				Memory:    data.Memory,
				HugePages: data.HugePages,
				Load1:     data.Load1,
				Load5:     data.Load5,
				Load15:    data.Load15,
			},
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

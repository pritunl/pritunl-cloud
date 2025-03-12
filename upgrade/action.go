package upgrade

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/instance"
)

func actionUpgrade(db *database.Database) (err error) {
	instances, err := instance.GetAll(db, &bson.M{})
	if err != nil {
		return
	}

	for _, inst := range instances {
		if inst.Action != "" || inst.State == instance.Active {
			continue
		}
		inst.Action = inst.State
		inst.State = instance.Active

		err = inst.CommitFields(db, set.NewSet("action", "state"))
		if err != nil {
			return
		}
	}

	return
}

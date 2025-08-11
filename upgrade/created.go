package upgrade

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/instance"
)

func createdUpgrade(db *database.Database) (err error) {
	insts, err := instance.GetAll(db, &bson.M{
		"created": &bson.M{
			"$exists": false,
		},
	})
	if err != nil {
		return
	}

	for _, inst := range insts {
		inst.Created = inst.Id.Timestamp()
		err = inst.CommitFields(db, set.NewSet("created"))
		if err != nil {
			return
		}
	}

	disks, err := disk.GetAll(db, &bson.M{
		"created": &bson.M{
			"$exists": false,
		},
	})
	if err != nil {
		return
	}

	for _, disk := range disks {
		disk.Created = disk.Id.Timestamp()
		err = disk.CommitFields(db, set.NewSet("created"))
		if err != nil {
			return
		}
	}

	return
}

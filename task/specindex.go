package task

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/spec"
	"github.com/pritunl/pritunl-cloud/unit"
)

var specIndex = &Task{
	Name:    "spec_index",
	Version: 1,
	Hours:   []int{6},
	Minutes: []int{32},
	Handler: specIndexHandler,
}

func specIndexSyncUnit(db *database.Database, unt *unit.Unit) (err error) {
	specs, err := spec.GetAllIndexes(db, &bson.M{
		"unit": unt.Id,
	})

	index := 0
	for i, spc := range specs {
		index = i + 1

		if spc.Index != index {
			spc.Index = index
			err = spc.CommitFields(db, set.NewSet("index"))
			if err != nil {
				return
			}
		}
	}

	if unt.SpecIndex != index {
		unt.SpecIndex = index
		err = unt.CommitFields(db, set.NewSet("spec_index"))
		if err != nil {
			return
		}
	}

	return
}

func specIndexHandler(db *database.Database) (err error) {
	units, err := unit.GetAll(db, &bson.M{})
	if err != nil {
		return
	}

	for _, unt := range units {
		err = specIndexSyncUnit(db, unt)
		if err != nil {
			return
		}
	}

	return
}

func init() {
	register(specIndex)
}

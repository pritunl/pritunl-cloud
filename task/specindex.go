package task

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/pod"
	"github.com/pritunl/pritunl-cloud/spec"
)

var specIndex = &Task{
	Name:    "spec_index",
	Version: 1,
	Hours:   []int{6},
	Minutes: []int{32},
	Handler: specIndexHandler,
}

func specIndexSyncUnit(db *database.Database, unit *pod.Unit) (err error) {
	specs, err := spec.GetAllIndexes(db, &bson.M{
		"unit": unit.Id,
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

	if unit.SpecIndex != index {
		unit.SpecIndex = index
		err = unit.CommitFields(db, set.NewSet("spec_index"))
		if err != nil {
			return
		}
	}

	return
}

func specIndexHandler(db *database.Database) (err error) {
	pods, err := pod.GetAll(db, &bson.M{})
	if err != nil {
		return
	}

	for _, pd := range pods {
		for _, unit := range pd.Units {
			unit.Pod = pd
			err = specIndexSyncUnit(db, unit)
			if err != nil {
				return
			}
		}
	}

	return
}

func init() {
	register(specIndex)
}

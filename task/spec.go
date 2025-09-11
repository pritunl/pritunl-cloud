package task

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/deployment"
	"github.com/pritunl/pritunl-cloud/spec"
	"github.com/sirupsen/logrus"
)

var specs = &Task{
	Name:    "specs",
	Version: 1,
	Hours: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12,
		13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23},
	Minutes: []int{0, 5, 10, 15, 20, 25, 30, 35, 40, 45, 50, 55},
	Handler: specsHandler,
}

func specsHandler(db *database.Database) (err error) {
	deplys, err := deployment.GetAll(db, &bson.M{})
	if err != nil {
		return
	}

	specIdsSet := set.NewSet()
	for _, deply := range deplys {
		if deply.Kind != deployment.Instance {
			continue
		}
		specIdsSet.Add(deply.Spec)
	}

	specIds := []bson.ObjectID{}
	for specId := range specIdsSet.Iter() {
		specIds = append(specIds, specId.(bson.ObjectID))
	}

	specs, err := spec.GetAll(db, &bson.M{
		"_id": &bson.M{
			"$in": specIds,
		},
	})
	if err != nil {
		return
	}

	for _, spec := range specs {
		errData, e := spec.Refresh(db)
		if e != nil || errData != nil {
			err = e

			logrus.WithFields(logrus.Fields{
				"spec_id":    spec.Id.Hex(),
				"error":      err,
				"error_data": errData,
			}).Error("deploy: Failed to refresh active spec")

			err = nil
			errData = nil
		}
	}

	return
}

func init() {
	register(specs)
}

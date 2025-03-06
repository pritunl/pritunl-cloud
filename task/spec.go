package task

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
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
	Minutes: []int{0, 10, 20, 30, 40, 50},
	Handler: specsHandler,
}

func specsHandler(db *database.Database) (err error) {
	deplys, err := deployment.GetAll(db, &bson.M{})
	if err != nil {
		return
	}

	specIdsSet := set.NewSet()
	for _, deply := range deplys {
		specIdsSet.Add(deply.Spec)
	}

	specIds := []primitive.ObjectID{}
	for specId := range specIdsSet.Iter() {
		specIds = append(specIds, specId.(primitive.ObjectID))
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

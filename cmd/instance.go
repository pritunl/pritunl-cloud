package cmd

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/sirupsen/logrus"
)

func StartInstance(name string) (err error) {
	db := database.GetDatabase()
	defer db.Close()

	instances, err := instance.GetAll(db, &bson.M{
		"name": name,
	})

	for _, inst := range instances {
		if inst.State != instance.Start {
			logrus.WithFields(logrus.Fields{
				"instance_id": inst.Id.Hex(),
			}).Info("cmd: Starting instance")

			inst.State = instance.Start
			err = inst.CommitFields(db, set.NewSet("state"))
			if err != nil {
				return
			}
		}
	}

	return
}

func StopInstance(name string) (err error) {
	db := database.GetDatabase()
	defer db.Close()

	instances, err := instance.GetAll(db, &bson.M{
		"name": name,
	})

	for _, inst := range instances {
		if inst.State != instance.Stop {
			logrus.WithFields(logrus.Fields{
				"instance_id": inst.Id.Hex(),
			}).Info("cmd: Stopping instance")

			inst.State = instance.Stop
			err = inst.CommitFields(db, set.NewSet("state"))
			if err != nil {
				return
			}
		}
	}

	return
}

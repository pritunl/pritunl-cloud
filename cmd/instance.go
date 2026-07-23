package cmd

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/config"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
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
		if inst.Action != instance.Start {
			logrus.WithFields(logrus.Fields{
				"instance_id": inst.Id.Hex(),
			}).Info("cmd: Starting instance")

			inst.Action = instance.Start
			err = inst.CommitFields(db, set.NewSet("action"))
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
		if inst.Action != instance.Stop {
			logrus.WithFields(logrus.Fields{
				"instance_id": inst.Id.Hex(),
			}).Info("cmd: Stopping instance")

			inst.Action = instance.Stop
			err = inst.CommitFields(db, set.NewSet("action"))
			if err != nil {
				return
			}
		}
	}

	return
}

func Shutdown() (err error) {
	db := database.GetDatabase()
	defer db.Close()

	nodeId, err := bson.ObjectIDFromHex(config.Config.NodeId)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "cmd: Failed to parse ObjectId"),
		}
		return
	}

	instances, err := instance.GetAll(db, &bson.M{
		"node": nodeId,
	})

	for _, inst := range instances {
		if inst.Action != instance.Stop {
			logrus.WithFields(logrus.Fields{
				"instance_id":   inst.Id.Hex(),
				"instance_name": inst.Name,
			}).Info("cmd: Stopping instance")

			inst.Action = instance.Stop
			err = inst.CommitFields(db, set.NewSet("action"))
			if err != nil {
				return
			}
		}
	}

	return
}

package task

import (
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/deployment"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/sirupsen/logrus"
)

var deployments = &Task{
	Name: "deployments",
	Hours: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12,
		13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23},
	Minutes: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12,
		13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25,
		26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38,
		39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51,
		52, 53, 54, 55, 56, 57, 58, 59},
	Seconds:    5 * time.Second,
	Handler:    deploymentsHandler,
	DebugNodes: []string{"web"},
}

func deploymentCheckInstance(db *database.Database,
	deply *deployment.Deployment) (err error) {

	if deply.State == deployment.Reserved {
		return
	}

	inst, err := instance.Get(db, deply.Instance)
	if err != nil {
		if _, ok := err.(*database.NotFoundError); ok {
			inst = nil
			err = nil
		} else {
			return
		}
	}

	if inst == nil {
		logrus.WithFields(logrus.Fields{
			"deployment": deply.Id.Hex(),
			"instance":   deply.Instance.Hex(),
			"service":    deply.Service.Hex(),
			"unit":       deply.Unit.Hex(),
		}).Info("scheduler: Removing deployment for missing instance")

		err = deployment.Remove(db, deply.Id)
		if err != nil {
			return
		}

		return
	}

	if inst.State == instance.Stop {
		logrus.WithFields(logrus.Fields{
			"deployment": deply.Id.Hex(),
			"instance":   deply.Instance.Hex(),
			"service":    deply.Service.Hex(),
			"unit":       deply.Unit.Hex(),
		}).Info("scheduler: Starting deployed instance")

		inst.State = instance.Start
		errData, e := inst.Validate(db)
		if e != nil {
			err = e
			return
		}

		if errData != nil {
			logrus.WithFields(logrus.Fields{
				"deployment":    deply.Id.Hex(),
				"instance":      deply.Instance.Hex(),
				"service":       deply.Service.Hex(),
				"unit":          deply.Unit.Hex(),
				"error_code":    errData.Error,
				"error_message": errData.Message,
			}).Error("scheduler: Validate instance failed")
			return
		}

		err = inst.CommitFields(db, set.NewSet("state"))
		if err != nil {
			return
		}
	}

	return
}

func deploymentsHandler(db *database.Database) (err error) {
	deplys, err := deployment.GetAll(db)
	if err != nil {
		return
	}

	for _, deply := range deplys {
		switch deply.Kind {
		case deployment.Instance:
			err = deploymentCheckInstance(db, deply)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"deployment": deply.Id.Hex(),
					"instance":   deply.Instance.Hex(),
					"service":    deply.Service.Hex(),
					"unit":       deply.Unit.Hex(),
					"error":      err,
				}).Error("scheduler: Failed to check instance deployment")
				err = nil
			}
			break
		}
	}

	return
}

func init() {
	register(deployments)
}

package task

import (
	"time"

	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/deployment"
	"github.com/pritunl/pritunl-cloud/scheduler"
	"github.com/pritunl/pritunl-cloud/service"
	"github.com/sirupsen/logrus"
)

var schedule = &Task{
	Name: "schedule",
	Hours: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12,
		13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23},
	Minutes: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12,
		13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25,
		26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38,
		39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51,
		52, 53, 54, 55, 56, 57, 58, 59},
	Seconds: 5 * time.Second,
	Handler: scheduleHandler,
}

func scheduleUnits(db *database.Database) (err error) {
	services, err := service.GetAll(db, &bson.M{})
	if err != nil {
		return
	}

	units := []*service.Unit{}
	for _, srvc := range services {
		for unit := range srvc.IterInstances() {
			units = append(units, unit)
		}
	}

	deploymentIds, err := deployment.GetAllActiveIds(db)
	if err != nil {
		return
	}

	for _, unit := range units {
		for _, deply := range unit.Deployments {
			if !deploymentIds.Contains(deply.Id) {
				logrus.WithFields(logrus.Fields{
					"service":    unit.Service.Id.Hex(),
					"unit":       unit.Id.Hex(),
					"deployment": deply.Id.Hex(),
				}).Info("deploy: Removing deployment")

				err = unit.RemoveDeployement(db, deply.Id)
				if err != nil {
					return
				}
			}
		}
	}

	for _, unit := range units {
		if len(unit.Deployments) >= unit.Count {
			continue
		}

		err = scheduler.Schedule(db, unit)
		if err != nil {
			return
		}
	}

	return
}

func scheduleHandler(db *database.Database) (err error) {
	err = scheduleUnits(db)
	if err != nil {
		return
	}

	schds, err := scheduler.GetAll(db)
	if err != nil {
		return
	}

	for _, schd := range schds {
		if schd.Consumed >= schd.Count {
			err = scheduler.Remove(db, schd.Id)
			if err != nil {
				return
			}
		}
	}

	return
}

func init() {
	register(schedule)
}

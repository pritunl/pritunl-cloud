package task

import (
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/deployment"
	"github.com/pritunl/pritunl-cloud/eval"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/plan"
	"github.com/pritunl/pritunl-cloud/service"
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

func deploymentBuildData(servc *service.Service, unit *service.Unit,
	inst *instance.Instance) (data eval.Data, err error) {

	dataStrct := plan.Data{
		Service: plan.Service{
			Name: servc.Name,
		},
		Unit: plan.Unit{
			Name:  unit.Name,
			Count: unit.Count,
		},
		Instance: plan.Instance{
			Name:      inst.Name,
			State:     inst.State,
			VirtState: inst.VirtState,
		},
	}

	data, err = dataStrct.Export()
	if err != nil {
		return
	}

	return
}

func deploymentSetInstanceState(db *database.Database,
	deply *deployment.Deployment, inst *instance.Instance,
	state string) (err error) {

	inst.State = state
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

	return
}

func deploymentCheckInstance(db *database.Database,
	servicesMap map[primitive.ObjectID]*service.Service,
	unitsMap map[primitive.ObjectID]*service.Unit,
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

	srvc := servicesMap[deply.Service]
	if srvc == nil {
		logrus.WithFields(logrus.Fields{
			"deployment": deply.Id.Hex(),
			"instance":   deply.Instance.Hex(),
			"service":    deply.Service.Hex(),
			"unit":       deply.Unit.Hex(),
		}).Info("scheduler: Failed to find service for deployment")

		// err = deployment.Remove(db, deply.Id)
		// if err != nil {
		// 	return
		// }

		return
	}

	unit := unitsMap[deply.Unit]
	if unit == nil {
		logrus.WithFields(logrus.Fields{
			"deployment": deply.Id.Hex(),
			"instance":   deply.Instance.Hex(),
			"service":    deply.Service.Hex(),
			"unit":       deply.Unit.Hex(),
		}).Info("scheduler: Failed to find unit for deployment")

		// err = deployment.Remove(db, deply.Id)
		// if err != nil {
		// 	return
		// }

		return
	}

	if unit.Kind != service.InstanceKind || unit.Instance == nil {
		return
	}

	pln, err := plan.Get(db, unit.Instance.Plan)
	if pln == nil {
		logrus.WithFields(logrus.Fields{
			"deployment": deply.Id.Hex(),
			"instance":   deply.Instance.Hex(),
			"service":    deply.Service.Hex(),
			"unit":       deply.Unit.Hex(),
		}).Info("scheduler: Failed to find plan for deployment")

		// err = deployment.Remove(db, deply.Id)
		// if err != nil {
		// 	return
		// }

		return
	}

	data, err := deploymentBuildData(srvc, unit, inst)
	if err != nil {
		return
	}

	var statement *plan.Statement
	action := ""
	ttl := 0
	for _, statement = range pln.Statements {
		action, ttl, err = eval.Eval(data, statement.Statement)
		if err != nil {
			return
		}

		if action != "" {
			break
		}
	}

	if action != "" {
		logrus.WithFields(logrus.Fields{
			"deployment": deply.Id.Hex(),
			"instance":   deply.Instance.Hex(),
			"service":    deply.Service.Hex(),
			"unit":       deply.Unit.Hex(),
			"statement":  statement.Statement,
			"ttl":        ttl,
			"action":     action,
		}).Info("scheduler: Handling plan action")

		switch action {
		case plan.Start:
			err = deploymentSetInstanceState(db, deply, inst, instance.Start)
			if err != nil {
				return
			}
			break
		case plan.Stop:
			err = deploymentSetInstanceState(db, deply, inst, instance.Stop)
			if err != nil {
				return
			}
			break
		case plan.Restart:
			err = deploymentSetInstanceState(db, deply, inst, instance.Restart)
			if err != nil {
				return
			}
			break
		case plan.Destroy:
			err = deploymentSetInstanceState(db, deply, inst, instance.Destroy)
			if err != nil {
				return
			}
			break
		default:
			logrus.WithFields(logrus.Fields{
				"deployment": deply.Id.Hex(),
				"instance":   deply.Instance.Hex(),
				"service":    deply.Service.Hex(),
				"unit":       deply.Unit.Hex(),
				"statement":  statement.Statement,
				"ttl":        ttl,
				"action":     action,
			}).Error("scheduler: Unknown plan action")
		}
	}

	return
}

func deploymentsHandler(db *database.Database) (err error) {
	deployments, err := deployment.GetAll(db)
	if err != nil {
		return
	}

	services, err := service.GetAll(db, &bson.M{})
	if err != nil {
		return
	}

	servicesMap := map[primitive.ObjectID]*service.Service{}
	unitsMap := map[primitive.ObjectID]*service.Unit{}

	for _, srvc := range services {
		servicesMap[srvc.Id] = srvc
		for _, unit := range srvc.Units {
			unitsMap[unit.Id] = unit
		}
	}

	for _, deply := range deployments {
		switch deply.Kind {
		case deployment.Instance:
			err = deploymentCheckInstance(db, servicesMap, unitsMap, deply)
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

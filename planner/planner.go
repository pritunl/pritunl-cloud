package planner

import (
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

type Planner struct {
	servicesMap map[primitive.ObjectID]*service.Service
}

func (p *Planner) setInstanceState(db *database.Database,
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

func (p *Planner) checkInstance(db *database.Database,
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

	srvc := p.servicesMap[deply.Service]
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

	unit := srvc.GetUnit(deply.Unit)
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

	data, err := buildEvalData(srvc, unit, inst)
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

		// TODO Finish ttl
		// TODO Perform action only once every 1+ min

		println("**************************************************")
		println(statement.Statement)
		println(action)
		println("**************************************************")

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
			err = p.setInstanceState(db, deply, inst, instance.Start)
			if err != nil {
				return
			}
			break
		case plan.Stop:
			err = p.setInstanceState(db, deply, inst, instance.Stop)
			if err != nil {
				return
			}
			break
		case plan.Restart:
			err = p.setInstanceState(db, deply, inst, instance.Restart)
			if err != nil {
				return
			}
			break
		case plan.Destroy:
			err = p.setInstanceState(db, deply, inst, instance.Destroy)
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

func (p *Planner) ApplyPlans(db *database.Database) (err error) {
	deployments, err := deployment.GetAll(db)
	if err != nil {
		return
	}

	services, err := service.GetAll(db, &bson.M{})
	if err != nil {
		return
	}

	p.servicesMap = map[primitive.ObjectID]*service.Service{}

	for _, srvc := range services {
		p.servicesMap[srvc.Id] = srvc
	}

	for _, deply := range deployments {
		switch deply.Kind {
		case deployment.Instance:
			err = p.checkInstance(db, deply)
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

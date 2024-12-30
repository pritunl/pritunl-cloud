package planner

import (
	"sync"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/deployment"
	"github.com/pritunl/pritunl-cloud/eval"
	"github.com/pritunl/pritunl-cloud/imds/types"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/plan"
	"github.com/pritunl/pritunl-cloud/service"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/spec"
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

	if inst == nil && deply.Kind == deployment.Instance {
		logrus.WithFields(logrus.Fields{
			"deployment": deply.Id.Hex(),
			"instance":   deply.Instance.Hex(),
			"service":    deply.Service.Hex(),
			"unit":       deply.Unit.Hex(),
		}).Info("scheduler: Removing deployment for destroyed instance")

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
		}).Error("scheduler: Failed to find service for deployment")

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
		}).Error("scheduler: Failed to find unit for deployment")

		// err = deployment.Remove(db, deply.Id)
		// if err != nil {
		// 	return
		// }

		return
	}

	if inst == nil {
		return
	}

	if deply.State == deployment.Archive && !inst.IsActive() {
		deply.State = deployment.Archived
		err = deply.CommitFields(db, set.NewSet("state"))
		if err != nil {
			return
		}
	}

	if deply.State == deployment.Archived && inst.State != instance.Stop {
		logrus.WithFields(logrus.Fields{
			"instance_id": inst.Id.Hex(),
		}).Info("deploy: Stopping instance for archived deployment")

		err = instance.SetState(db, inst.Id, instance.Stop)
		if err != nil {
			return
		}
	}

	if deply.State == deployment.Restore && inst.IsActive() {
		deply.State = deployment.Deployed
		err = deply.CommitFields(db, set.NewSet("state"))
		if err != nil {
			return
		}
	}

	status := deployment.Unhealthy
	if inst.Guest != nil {
		heartbeatTtl := time.Duration(
			settings.System.InstanceTimestampTtl) * time.Second
		if inst.Guest.Status == types.Running &&
			time.Since(inst.Guest.Heartbeat) <= heartbeatTtl {

			status = deployment.Healthy
		}
	}

	if deply.Status != status {
		deply.Status = status
		err = deply.CommitFields(db, set.NewSet("status"))
		if err != nil {
			return
		}
	}

	switch deply.State {
	case deployment.Archive, deployment.Archived, deployment.Restore:
		return
	}

	if deply.Kind != deployment.Image &&
		deply.State == deployment.Deployed &&
		!inst.IsActive() {

		logrus.WithFields(logrus.Fields{
			"instance_id": inst.Id.Hex(),
		}).Info("deploy: Starting instance for active deployment")

		err = instance.SetState(db, inst.Id, instance.Start)
		if err != nil {
			return
		}

		return
	}

	if deply.State == deployment.Deployed && !unit.HasDeployment(deply.Id) {
		logrus.WithFields(logrus.Fields{
			"deployment": deply.Id.Hex(),
			"instance":   deply.Instance.Hex(),
			"service":    deply.Service.Hex(),
			"unit":       deply.Unit.Hex(),
		}).Info("scheduler: Restoring deployment")

		err = unit.RestoreDeployment(db, deply.Id)
		if err != nil {
			return
		}
	}

	spc, err := spec.Get(db, deply.Spec)
	if err != nil {
		return
	}

	if spc.Instance == nil {
		return
	}

	pln, err := plan.Get(db, spc.Instance.Plan)
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
	threshold := 0
	for _, statement = range pln.Statements {
		action, threshold, err = eval.Eval(data, statement.Statement)
		if err != nil {
			return
		}

		log := false
		if action != "" {
			log = true
			println("**************************************************")
			println(action)
			println(threshold)
		}

		action, err = deply.HandleStatement(
			db, statement.Id, threshold, action)
		if err != nil {
			return
		}

		if log {
			println(action)
			println("**************************************************")
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
			"threshold":  threshold,
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
				"threshold":  threshold,
				"action":     action,
			}).Error("scheduler: Unknown plan action")
		}
	}

	return
}

func (p *Planner) ApplyPlans(db *database.Database) (err error) {
	deployments, err := deployment.GetAll(db, &bson.M{})
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

	var waiters sync.WaitGroup
	batch := make(chan struct{}, settings.System.PlannerBatchSize)

	for _, deply := range deployments {
		waiters.Add(1)
		batch <- struct{}{}

		go func(deply *deployment.Deployment) {
			defer func() {
				<-batch
				waiters.Done()
			}()

			switch deply.Kind {
			case deployment.Instance, deployment.Image:
				e := p.checkInstance(db, deply)
				if e != nil {
					logrus.WithFields(logrus.Fields{
						"deployment": deply.Id.Hex(),
						"instance":   deply.Instance.Hex(),
						"service":    deply.Service.Hex(),
						"unit":       deply.Unit.Hex(),
						"error":      e,
					}).Error("scheduler: Failed to check instance deployment")
				}
				break
			}
		}(deply)
	}

	waiters.Wait()
	return
}

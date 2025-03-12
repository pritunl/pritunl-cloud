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
	"github.com/pritunl/pritunl-cloud/pod"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/spec"
	"github.com/sirupsen/logrus"
)

type Planner struct {
	podsMap map[primitive.ObjectID]*pod.Pod
}

func (p *Planner) setInstanceAction(db *database.Database,
	deply *deployment.Deployment, inst *instance.Instance,
	action string) (err error) {

	inst.Action = action
	errData, e := inst.Validate(db)
	if e != nil {
		err = e
		return
	}

	if errData != nil {
		logrus.WithFields(logrus.Fields{
			"deployment":    deply.Id.Hex(),
			"instance":      deply.Instance.Hex(),
			"pod":           deply.Pod.Hex(),
			"unit":          deply.Unit.Hex(),
			"error_code":    errData.Error,
			"error_message": errData.Message,
		}).Error("scheduler: Validate instance failed")
		return
	}

	err = inst.CommitFields(db, set.NewSet("action"))
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
			"pod":        deply.Pod.Hex(),
			"unit":       deply.Unit.Hex(),
		}).Info("scheduler: Removing deployment for destroyed instance")

		err = deployment.Remove(db, deply.Id)
		if err != nil {
			return
		}

		return
	}

	pd := p.podsMap[deply.Pod]
	if pd == nil {
		logrus.WithFields(logrus.Fields{
			"deployment": deply.Id.Hex(),
			"instance":   deply.Instance.Hex(),
			"pod":        deply.Pod.Hex(),
			"unit":       deply.Unit.Hex(),
		}).Error("scheduler: Failed to find pod for deployment")

		// err = deployment.Remove(db, deply.Id)
		// if err != nil {
		// 	return
		// }

		return
	}

	unit := pd.GetUnit(deply.Unit)
	if unit == nil {
		logrus.WithFields(logrus.Fields{
			"deployment": deply.Id.Hex(),
			"instance":   deply.Instance.Hex(),
			"pod":        deply.Pod.Hex(),
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

	if deply.State == deployment.Archived && inst.Action != instance.Stop {
		logrus.WithFields(logrus.Fields{
			"instance_id": inst.Id.Hex(),
		}).Info("deploy: Stopping instance for archived deployment")

		err = instance.SetAction(db, inst.Id, instance.Stop)
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

		err = instance.SetAction(db, inst.Id, instance.Start)
		if err != nil {
			return
		}

		return
	}

	if deply.State == deployment.Deployed && !unit.HasDeployment(deply.Id) {
		logrus.WithFields(logrus.Fields{
			"deployment": deply.Id.Hex(),
			"instance":   deply.Instance.Hex(),
			"pod":        deply.Pod.Hex(),
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
			"pod":        deply.Pod.Hex(),
			"unit":       deply.Unit.Hex(),
		}).Info("scheduler: Failed to find plan for deployment")

		// err = deployment.Remove(db, deply.Id)
		// if err != nil {
		// 	return
		// }

		return
	}

	data, err := buildEvalData(pd, unit, inst)
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
			"pod":        deply.Pod.Hex(),
			"unit":       deply.Unit.Hex(),
			"statement":  statement.Statement,
			"threshold":  threshold,
			"action":     action,
		}).Info("scheduler: Handling plan action")

		switch action {
		case plan.Start:
			err = p.setInstanceAction(db, deply, inst, instance.Start)
			if err != nil {
				return
			}
			break
		case plan.Stop:
			err = p.setInstanceAction(db, deply, inst, instance.Stop)
			if err != nil {
				return
			}
			break
		case plan.Restart:
			err = p.setInstanceAction(db, deply, inst, instance.Restart)
			if err != nil {
				return
			}
			break
		case plan.Destroy:
			err = p.setInstanceAction(db, deply, inst, instance.Destroy)
			if err != nil {
				return
			}
			break
		default:
			logrus.WithFields(logrus.Fields{
				"deployment": deply.Id.Hex(),
				"instance":   deply.Instance.Hex(),
				"pod":        deply.Pod.Hex(),
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

	pods, err := pod.GetAll(db, &bson.M{})
	if err != nil {
		return
	}

	p.podsMap = map[primitive.ObjectID]*pod.Pod{}

	for _, pd := range pods {
		p.podsMap[pd.Id] = pd
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
						"pod":        deply.Pod.Hex(),
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

package planner

import (
	"sync"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/deployment"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/eval"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/imds/types"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/plan"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/spec"
	"github.com/pritunl/pritunl-cloud/unit"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/sirupsen/logrus"
)

type Planner struct {
	unitsMap map[primitive.ObjectID]*unit.Unit
}

func (p *Planner) setInstanceAction(db *database.Database,
	deply *deployment.Deployment, inst *instance.Instance,
	statement *plan.Statement, threshold int, action string) (err error) {

	disks, e := disk.GetInstance(db, inst.Id)
	if e != nil {
		err = e
		return
	}

	for _, dsk := range disks {
		if dsk.Action != "" {
			logrus.WithFields(logrus.Fields{
				"instance_id": inst.Id.Hex(),
				"disk_id":     dsk.Id.Hex(),
				"disk_action": dsk.Action,
			}).Info("deploy: Ignoring instance plan action, " +
				"disk action pending")
			return
		}
	}

	if inst.Action == action {
		return
	}

	logrus.WithFields(logrus.Fields{
		"deployment": deply.Id.Hex(),
		"instance":   deply.Instance.Hex(),
		"pod":        deply.Pod.Hex(),
		"unit":       deply.Unit.Hex(),
		"statement":  statement.Statement,
		"threshold":  threshold,
		"action":     action,
	}).Info("scheduler: Handling plan action")

	inst.Action = action
	errData, e := inst.Validate(db)
	if e != nil {
		err = e
		return
	}

	if errData != nil {
		err = errData.GetError()
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

	unt := p.unitsMap[deply.Unit]
	if unt == nil {
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

	if deply.Action == deployment.Restore && inst.IsActive() {
		deply.Action = ""
		deply.State = deployment.Deployed
		err = deply.CommitFields(db, set.NewSet("state", "action"))
		if err != nil {
			return
		}
	}

	status := deployment.Unhealthy
	if inst.Guest != nil {
		if inst.Guest.Status == types.Running ||
			inst.Guest.Status == types.ReloadingClean {

			now := time.Now()
			heartbeatTtl := time.Duration(
				settings.System.InstanceTimestampTtl) * time.Second

			if now.Sub(inst.Guest.Heartbeat) <= heartbeatTtl {
				status = deployment.Healthy
			} else if now.Sub(inst.Guest.Timestamp) > heartbeatTtl {
				status = deployment.Unknown
			}
		}
	}

	if deply.Status != status {
		deply.Status = status
		err = deply.CommitFields(db, set.NewSet("status"))
		if err != nil {
			return
		}
	}

	if deply.Action != "" {
		return
	}

	switch deply.State {
	case deployment.Archived:
		return
	}

	if deply.State == deployment.Deployed && !unt.HasDeployment(deply.Id) {
		logrus.WithFields(logrus.Fields{
			"deployment": deply.Id.Hex(),
			"instance":   deply.Instance.Hex(),
			"pod":        deply.Pod.Hex(),
			"unit":       deply.Unit.Hex(),
		}).Info("scheduler: Restoring deployment")

		err = unt.RestoreDeployment(db, deply.Id)
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

	if deply.State != deployment.Deployed {
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
		return
	}

	data, err := buildEvalData(unt, inst)
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
		threshold = utils.Max(deployment.ThresholdMin, threshold)

		action, err = deply.HandleStatement(
			db, statement.Id, threshold, action)
		if err != nil {
			return
		}

		if action != "" {
			break
		}
	}

	if action != "" {
		switch action {
		case plan.Start:
			err = p.setInstanceAction(db, deply, inst,
				statement, threshold, instance.Start)
			if err != nil {
				return
			}
			break
		case plan.Stop:
			err = p.setInstanceAction(db, deply, inst,
				statement, threshold, instance.Stop)
			if err != nil {
				return
			}
			break
		case plan.Restart:
			err = p.setInstanceAction(db, deply, inst,
				statement, threshold, instance.Restart)
			if err != nil {
				return
			}
			break
		case plan.Destroy:
			err = p.setInstanceAction(db, deply, inst,
				statement, threshold, instance.Destroy)
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

	p.unitsMap, err = unit.GetAllMap(db, &bson.M{})
	if err != nil {
		return
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

			if deply.State == deployment.Reserved &&
				deply.Action == deployment.Destroy &&
				time.Since(deply.Timestamp) > 300*time.Second {

				err := deployment.Remove(db, deply.Id)
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"deployment_id": deply.Id.Hex(),
						"error":         err,
					}).Error("deploy: Failed to remove deployment")
					return
				}

				event.PublishDispatch(db, "pod.change")
			}
		}(deply)
	}

	waiters.Wait()
	return
}

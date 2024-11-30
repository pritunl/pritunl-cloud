package deploy

import (
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/deployment"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/scheduler"
	"github.com/pritunl/pritunl-cloud/service"
	"github.com/pritunl/pritunl-cloud/spec"
	"github.com/pritunl/pritunl-cloud/state"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/sirupsen/logrus"
)

var (
	servicesLock    = utils.NewMultiTimeoutLock(3 * time.Minute)
	servicesLimiter = utils.NewLimiter(50)
)

type Services struct {
	stat *state.State
}

func (s *Services) processSchedule(schd *scheduler.Scheduler) {
	if !servicesLimiter.Acquire() {
		return
	}

	acquired, lockId := instancesLock.LockOpen(schd.Id.Unit.Hex())
	if !acquired {
		return
	}

	go func() {
		defer func() {
			time.Sleep(1 * time.Second)
			instancesLock.Unlock(schd.Id.Unit.Hex(), lockId)
			servicesLimiter.Release()
		}()

		err := s.deploySchedule(schd)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"service": schd.Id.Service.Hex(),
				"unit":    schd.Id.Unit.Hex(),
				"error":   err,
			}).Error("deploy: Service deploy failed")
			return
		}
	}()
}

func (s *Services) deploySchedule(schd *scheduler.Scheduler) (err error) {
	db := database.GetDatabase()
	defer db.Close()

	servc, err := service.Get(db, schd.Id.Service)
	if err != nil {
		return
	}

	unit := servc.GetUnit(schd.Id.Unit)
	if unit == nil {
		logrus.WithFields(logrus.Fields{
			"service": schd.Id.Service.Hex(),
			"unit":    schd.Id.Unit.Hex(),
		}).Info("deploy: Service deploy nil unit")
		return
	}

	spc, err := spec.Get(db, schd.Spec)
	if err != nil {
		return
	}

	tickets := schd.Tickets[s.stat.Node().Id]
	if tickets != nil && len(tickets) > 0 {
		now := time.Now()
		for _, ticket := range tickets {
			start := schd.Created.Add(
				time.Duration(ticket.Offset) * time.Second)
			if now.After(start) {
				exists, e := schd.Refresh(db)
				if e != nil {
					err = e
					return
				}

				if !exists {
					logrus.WithFields(logrus.Fields{
						"service": schd.Id.Service.Hex(),
						"unit":    schd.Id.Unit.Hex(),
					}).Info("deploy: Service deploy schedule lost")
					return
				}

				if schd.Consumed >= schd.Count {
					return
				}

				err = s.DeploySpec(db, schd, unit, spc)
				if err != nil {
					return
				}

				err = schd.Consume(db)
				if err != nil {
					return
				}
			}
		}
	}

	return
}

func (s *Services) DeploySpec(db *database.Database,
	schd *scheduler.Scheduler, unit *service.Unit,
	spc *spec.Commit) (err error) {

	deply := &deployment.Deployment{
		Service: unit.Service.Id,
		Unit:    unit.Id,
		Spec:    spc.Id,
		Node:    node.Self.Id,
		Kind:    unit.Kind,
		State:   deployment.Reserved,
	}

	errData, err := deply.Validate(db)
	if err != nil {
		return
	}

	if errData != nil {
		logrus.WithFields(logrus.Fields{
			"error_code":    errData.Error,
			"error_message": errData.Message,
		}).Error("deploy: Failed to validate deployment")
		return
	}

	err = deply.Insert(db)
	if err != nil {
		return
	}

	reserved, err := unit.Reserve(db, deply.Id, schd.OverrideCount)
	if err != nil {
		return
	}

	if !reserved {
		err = deployment.Remove(db, deply.Id)
		if err != nil {
			return
		}
		return
	}

	inst := &instance.Instance{
		Organization: unit.Service.Organization,
		Zone:         spc.Instance.Zone,
		Vpc:          spc.Instance.Vpc,
		Subnet:       spc.Instance.Subnet,
		//Shape:               spc.Instance.Shape,
		Node:                node.Self.Id,
		Image:               spc.Instance.Image,
		Uefi:                true,
		SecureBoot:          true,
		Tpm:                 false,
		DhcpServer:          false,
		CloudType:           instance.Linux,
		CloudScript:         "",
		DeleteProtection:    false,
		SkipSourceDestCheck: false,
		Name:                spc.Name,
		Comment:             "",
		InitDiskSize:        10,
		Memory:              2048,
		Processors:          2,
		NetworkRoles:        spc.Instance.Roles,
		NoPublicAddress:     false,
		NoPublicAddress6:    false,
		NoHostAddress:       false,
		Deployment:          deply.Id,
	}

	errData, err = inst.Validate(db)
	if err != nil {
		return
	}

	if errData != nil {
		logrus.WithFields(logrus.Fields{
			"error_code":    errData.Error,
			"error_message": errData.Message,
		}).Error("deploy: Failed to deploy instance")
		return
	}

	err = inst.Insert(db)
	if err != nil {
		return
	}

	deply.State = deployment.Deployed
	deply.Instance = inst.Id
	err = deply.CommitFields(db, set.NewSet("state", "instance"))
	if err != nil {
		return
	}

	event.PublishDispatch(db, "service.change")

	return
}

func (s *Services) Deploy(db *database.Database) (err error) {
	schds := s.stat.Schedulers()

	for _, schd := range schds {
		if schd.Kind != scheduler.InstanceUnitKind {
			continue
		}

		tickets := schd.Tickets[s.stat.Node().Id]
		if tickets != nil && len(tickets) > 0 {
			now := time.Now()
			for _, ticket := range tickets {
				start := schd.Created.Add(
					time.Duration(ticket.Offset) * time.Second)
				if now.After(start) {
					s.processSchedule(schd)
					break
				}
			}
		}
	}

	return
}

func NewServices(stat *state.State) *Services {
	return &Services{
		stat: stat,
	}
}

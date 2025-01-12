package deploy

import (
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/deployment"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/pod"
	"github.com/pritunl/pritunl-cloud/scheduler"
	"github.com/pritunl/pritunl-cloud/spec"
	"github.com/pritunl/pritunl-cloud/state"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/sirupsen/logrus"
)

var (
	podsLock    = utils.NewMultiTimeoutLock(3 * time.Minute)
	podsLimiter = utils.NewLimiter(50)
)

type Pods struct {
	stat *state.State
}

func (s *Pods) processSchedule(schd *scheduler.Scheduler) {
	if !podsLimiter.Acquire() {
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
			podsLimiter.Release()
		}()

		err := s.deploySchedule(schd)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"pod":   schd.Id.Pod.Hex(),
				"unit":  schd.Id.Unit.Hex(),
				"error": err,
			}).Error("deploy: Pod deploy failed")
			return
		}
	}()
}

func (s *Pods) deploySchedule(schd *scheduler.Scheduler) (err error) {
	db := database.GetDatabase()
	defer db.Close()

	pd, err := pod.Get(db, schd.Id.Pod)
	if err != nil {
		return
	}

	unit := pd.GetUnit(schd.Id.Unit)
	if unit == nil {
		logrus.WithFields(logrus.Fields{
			"pod":  schd.Id.Pod.Hex(),
			"unit": schd.Id.Unit.Hex(),
		}).Info("deploy: Pod deploy nil unit")
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
						"pod":  schd.Id.Pod.Hex(),
						"unit": schd.Id.Unit.Hex(),
					}).Info("deploy: Pod deploy schedule lost")
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

func (s *Pods) DeploySpec(db *database.Database,
	schd *scheduler.Scheduler, unit *pod.Unit,
	spc *spec.Commit) (err error) {

	deply := &deployment.Deployment{
		Pod:       unit.Pod.Id,
		Unit:      unit.Id,
		Timestamp: time.Now(),
		Spec:      spc.Id,
		Zone:      node.Self.Zone,
		Node:      node.Self.Id,
		Kind:      unit.Kind,
		State:     deployment.Reserved,
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
		Organization: unit.Pod.Organization,
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

	inst.GenerateId()

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

	index := 0
	reservedDisks := []*disk.Disk{}
	for _, mount := range spc.Instance.Mounts {
		index += 1
		reserved := false

		for _, dskId := range mount.Disks {
			dsk, e := disk.Get(db, dskId)
			if e != nil {
				err = e
				return
			}

			if !dsk.Instance.IsZero() {
				continue
			}

			err = dsk.Reserve(db, inst.Id, index)
			if err != nil {
				return
			}

			reserved = true
			reservedDisks = append(reservedDisks, dsk)
			break
		}

		if !reserved {
			for _, dsk := range reservedDisks {
				err = dsk.Unreserve(db, inst.Id)
				if err != nil {
					return
				}
			}

			logrus.WithFields(logrus.Fields{
				"mount_path": mount.Path,
			}).Error("deploy: Failed to reserve disk for mount")
			return
		}
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

	event.PublishDispatch(db, "pod.change")

	return
}

func (s *Pods) Deploy(db *database.Database) (err error) {
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

func NewPods(stat *state.State) *Pods {
	return &Pods{
		stat: stat,
	}
}

package deploy

import (
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/deployment"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/image"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/nodeport"
	"github.com/pritunl/pritunl-cloud/scheduler"
	"github.com/pritunl/pritunl-cloud/spec"
	"github.com/pritunl/pritunl-cloud/state"
	"github.com/pritunl/pritunl-cloud/unit"
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

	acquired, lockId := podsLock.LockOpen(schd.Id.Hex())
	if !acquired {
		return
	}

	go func() {
		defer func() {
			time.Sleep(1 * time.Second)
			podsLock.Unlock(schd.Id.Hex(), lockId)
			podsLimiter.Release()
		}()

		err := s.deploySchedule(schd)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"unit":  schd.Id.Hex(),
				"error": err,
			}).Error("deploy: Unit deploy failed")
			return
		}
	}()
}

func (s *Pods) deploySchedule(schd *scheduler.Scheduler) (err error) {
	db := database.GetDatabase()
	defer db.Close()

	unt, err := unit.Get(db, schd.Id)
	if err != nil {
		return
	}

	spc, err := spec.Get(db, schd.Spec)
	if err != nil {
		return
	}

	tickets := schd.Tickets[s.stat.Node().Id]
	if len(tickets) > 0 {
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
						"pod":  unt.Pod.Hex(),
						"unit": unt.Id.Hex(),
					}).Info("deploy: Pod deploy schedule lost")
					return
				}

				if schd.Consumed >= schd.Count {
					return
				}

				if !schd.Ready() {
					logrus.WithFields(logrus.Fields{
						"pod":  unt.Pod.Hex(),
						"unit": unt.Id.Hex(),
					}).Info("deploy: Reached maximum schedule attempts")

					err = schd.ClearTickets(db)
					if err != nil {
						return
					}
					return
				}

				reserved, e := s.DeploySpec(db, schd, unt, spc)
				if e != nil {
					err = e

					limit, _ := schd.Failure(db)
					if limit {
						logrus.WithFields(logrus.Fields{
							"pod":  unt.Pod.Hex(),
							"unit": unt.Id.Hex(),
						}).Info("deploy: Reached maximum schedule attempts")
					}
					return
				}

				if reserved {
					err = schd.Consume(db)
					if err != nil {
						return
					}
				} else {
					limit, e := schd.Failure(db)
					if e != nil {
						err = e
						return
					}

					if limit {
						logrus.WithFields(logrus.Fields{
							"pod":  unt.Pod.Hex(),
							"unit": unt.Id.Hex(),
						}).Info("deploy: Reached maximum schedule attempts")
					}
				}
			}
		}
	}

	return
}

func (s *Pods) DeploySpec(db *database.Database,
	schd *scheduler.Scheduler, unt *unit.Unit,
	spc *spec.Spec) (reserved bool, err error) {

	img, err := image.Get(db, spc.Instance.Image)
	if err != nil {
		return
	}

	deply := &deployment.Deployment{
		Pod:          unt.Pod,
		Unit:         unt.Id,
		Organization: unt.Organization,
		Timestamp:    time.Now(),
		Spec:         spc.Id,
		Datacenter:   node.Self.Datacenter,
		Zone:         node.Self.Zone,
		Node:         node.Self.Id,
		Kind:         unt.Kind,
		State:        deployment.Reserved,
	}

	errData, err := spc.Refresh(db)
	if err != nil {
		return
	}

	if errData != nil {
		err = errData.GetError()
		return
	}

	errData, err = deply.Validate(db)
	if err != nil {
		return
	}

	if errData != nil {
		err = errData.GetError()
		return
	}

	err = deply.Insert(db)
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			e := deployment.Remove(db, deply.Id)
			if e != nil {
				logrus.WithFields(logrus.Fields{
					"error": e,
				}).Error("deploy: Failed to cleanup deployment")
				return
			}
		}
	}()

	err = unt.Refresh(db)
	if err != nil {
		return
	}

	reserved, err = unt.Reserve(db, deply.Id, schd.OverrideCount)
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
		Organization:        unt.Organization,
		Zone:                spc.Instance.Zone,
		Vpc:                 spc.Instance.Vpc,
		Subnet:              spc.Instance.Subnet,
		Shape:               spc.Instance.Shape,
		Node:                node.Self.Id,
		Image:               spc.Instance.Image,
		Uefi:                true,
		Tpm:                 spc.Instance.Tpm,
		Vnc:                 spc.Instance.Vnc,
		DhcpServer:          spc.Instance.DhcpServer,
		CloudScript:         "",
		DeleteProtection:    spc.Instance.DeleteProtection,
		SkipSourceDestCheck: spc.Instance.SkipSourceDestCheck,
		Name:                spc.Name,
		Comment:             "",
		InitDiskSize:        10,
		Memory:              spc.Instance.Memory,
		Processors:          spc.Instance.Processors,
		NetworkRoles:        spc.Instance.Roles,
		NoPublicAddress:     false,
		NoPublicAddress6:    false,
		NoHostAddress:       false,
		Deployment:          deply.Id,
	}

	if img.GetSystemType() == image.Bsd {
		inst.CloudType = instance.BSD
		inst.SecureBoot = false
	} else {
		inst.CloudType = instance.Linux
		inst.SecureBoot = true
	}

	if spc.Instance.Uefi != nil {
		inst.Uefi = *spc.Instance.Uefi
	}
	if spc.Instance.SecureBoot != nil {
		inst.SecureBoot = *spc.Instance.SecureBoot
	}
	if spc.Instance.CloudType != "" {
		inst.CloudType = spc.Instance.CloudType
	}
	if spc.Instance.HostAddress != nil {
		inst.NoHostAddress = !*spc.Instance.HostAddress
	}
	if spc.Instance.PublicAddress != nil {
		inst.NoPublicAddress = !*spc.Instance.PublicAddress
	} else {
		inst.NoPublicAddress = node.Self.DefaultNoPublicAddress
	}
	if spc.Instance.PublicAddress6 != nil {
		inst.NoPublicAddress6 = !*spc.Instance.PublicAddress6
	} else {
		inst.NoPublicAddress6 = node.Self.DefaultNoPublicAddress6
	}
	if spc.Instance.DiskSize != 0 {
		inst.InitDiskSize = spc.Instance.DiskSize
	}

	if len(spc.Instance.NodePorts) > 0 {
		for _, ndePort := range spc.Instance.NodePorts {
			inst.NodePorts = append(inst.NodePorts, &nodeport.Mapping{
				Protocol:     ndePort.Protocol,
				ExternalPort: ndePort.ExternalPort,
				InternalPort: ndePort.InternalPort,
			})
		}
	}

	err = inst.GenerateId()
	if err != nil {
		return
	}

	errData, err = inst.Validate(db)
	if err != nil {
		return
	}

	if errData != nil {
		reserved = false
		err = errData.GetError()
		return
	}

	if len(inst.NodePorts) > 0 {
		err = inst.SyncNodePorts(db)
		if err != nil {
			return
		}
	}

	index := 0
	reservedDisks := []*disk.Disk{}
	deplyMounts := []*deployment.Mount{}

	for _, mount := range spc.Instance.Mounts {
		index += 1
		diskReserved := false

		for _, dskId := range mount.Disks {
			dsk, e := disk.Get(db, dskId)
			if e != nil {
				err = e

				for _, dsk := range reservedDisks {
					err = dsk.Unreserve(db, inst.Id, deply.Id)
					if err != nil {
						return
					}
				}

				return
			}

			if dsk.Node != node.Self.Id || !dsk.Instance.IsZero() {
				continue
			}

			diskReserved, err = dsk.Reserve(db, inst.Id, index, deply.Id)
			if err != nil {
				for _, dsk := range reservedDisks {
					err = dsk.Unreserve(db, inst.Id, deply.Id)
					if err != nil {
						return
					}
				}
				return
			}

			if !diskReserved {
				continue
			}

			deplyMounts = append(deplyMounts, &deployment.Mount{
				Disk: dsk.Id,
				Path: mount.Path,
				Uuid: dsk.Uuid,
			})

			reservedDisks = append(reservedDisks, dsk)
			break
		}

		if !diskReserved {
			for _, dsk := range reservedDisks {
				err = dsk.Unreserve(db, inst.Id, deply.Id)
				if err != nil {
					return
				}
			}

			logrus.WithFields(logrus.Fields{
				"mount_path": mount.Path,
			}).Error("deploy: Failed to reserve disk for mount")

			err = deployment.Remove(db, deply.Id)
			if err != nil {
				return
			}

			reserved = false

			return
		}
	}

	err = inst.Insert(db)
	if err != nil {
		_ = inst.Cleanup(db)

		for _, dsk := range reservedDisks {
			err = dsk.Unreserve(db, inst.Id, deply.Id)
			if err != nil {
				return
			}
		}

		return
	}

	err = inst.Cleanup(db)
	if err != nil {
		return
	}

	deply.State = deployment.Deployed
	deply.Instance = inst.Id
	deply.Mounts = deplyMounts

	err = deply.CommitFields(db, set.NewSet("state", "instance", "mounts"))
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

		if len(schd.Tickets) == 0 {
			deleted, e := scheduler.Remove(db, schd.Id)
			if e != nil {
				err = e
				return
			}

			if deleted {
				logrus.WithFields(logrus.Fields{
					"unit": schd.Id.Hex(),
				}).Error("deploy: All nodes failed to schedule deployment")
			}
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

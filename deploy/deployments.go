package deploy

import (
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/deployment"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/domain"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/image"
	"github.com/pritunl/pritunl-cloud/imds"
	"github.com/pritunl/pritunl-cloud/imds/types"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/shape"
	"github.com/pritunl/pritunl-cloud/spec"
	"github.com/pritunl/pritunl-cloud/state"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/sirupsen/logrus"
)

var (
	deploymentsLock = utils.NewMultiTimeoutLock(5 * time.Minute)
)

type Deployments struct {
	stat *state.State
}

func (d *Deployments) migrate(db *database.Database,
	deply *deployment.Deployment) {

	inst := d.stat.GetInstace(deply.Instance)
	nodeId := d.stat.Node().Id

	acquired, lockId := deploymentsLock.LockOpenTimeout(
		deply.Id.Hex(), 3*time.Minute)
	if !acquired {
		return
	}

	go func() {
		defer func() {
			deploymentsLock.Unlock(deply.Id.Hex(), lockId)
		}()

		if deply.Node != nodeId {
			return
		}

		curSpec, err := spec.Get(db, deply.Spec)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"deployment_id": deply.Id.Hex(),
				"cur_spec_id":   curSpec.Id.Hex(),
				"new_spec_id":   deply.NewSpec.Hex(),
				"error":         err,
			}).Error("deploy: Failed to get current spec")
			return
		}

		newSpec, err := spec.Get(db, deply.NewSpec)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"deployment_id": deply.Id.Hex(),
				"cur_spec_id":   curSpec.Id.Hex(),
				"new_spec_id":   newSpec.Id.Hex(),
				"error":         err,
			}).Error("deploy: Failed to get new spec")
			return
		}

		errData, err := curSpec.CanMigrate(db, newSpec)
		if err != nil || errData != nil {
			logrus.WithFields(logrus.Fields{
				"deployment_id": deply.Id.Hex(),
				"cur_spec_id":   curSpec.Id.Hex(),
				"new_spec_id":   newSpec.Id.Hex(),
				"error":         err,
				"error_data":    errData,
			}).Error("deploy: Incompatible migrate")

			deply.State = deployment.Deployed
			deply.NewSpec = primitive.NilObjectID
			err = deply.CommitFields(db, set.NewSet("state", "new_spec"))
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"deployment_id": deply.Id.Hex(),
					"cur_spec_id":   curSpec.Id.Hex(),
					"new_spec_id":   newSpec.Id.Hex(),
					"error":         err,
				}).Error("deploy: Failed to commit deployment")
				return
			}

			return
		}

		if curSpec.Instance.Processors != newSpec.Instance.Processors ||
			curSpec.Instance.Memory != newSpec.Instance.Memory {

			flexible := true
			if !newSpec.Instance.Shape.IsZero() {
				shp, e := shape.Get(db, newSpec.Instance.Shape)
				if e != nil {
					err = e

					logrus.WithFields(logrus.Fields{
						"deployment_id": deply.Id.Hex(),
						"cur_spec_id":   curSpec.Id.Hex(),
						"new_spec_id":   newSpec.Id.Hex(),
						"error":         err,
					}).Error("deploy: Failed to get spec shape")
					return
				}

				flexible = shp.Flexible
			}

			if flexible && inst != nil {
				inst.Processors = newSpec.Instance.Processors
				inst.Memory = newSpec.Instance.Memory

				err = inst.CommitFields(
					db, set.NewSet("processors", "memory"))
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"deployment_id": deply.Id.Hex(),
						"cur_spec_id":   curSpec.Id.Hex(),
						"new_spec_id":   newSpec.Id.Hex(),
						"error":         err,
					}).Error("deploy: Failed to migrate instance")
					return
				}
			}
		}

		deply.State = deployment.Deployed
		deply.Spec = newSpec.Id
		deply.NewSpec = primitive.NilObjectID
		err = deply.CommitFields(db, set.NewSet("state", "spec", "new_spec"))
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"deployment_id": deply.Id.Hex(),
				"cur_spec_id":   curSpec.Id.Hex(),
				"new_spec_id":   newSpec.Id.Hex(),
				"error":         err,
			}).Error("deploy: Failed to commit deployment")
			return
		}

		logrus.WithFields(logrus.Fields{
			"deployment_id": deply.Id.Hex(),
			"cur_spec_id":   curSpec.Id.Hex(),
			"new_spec_id":   newSpec.Id.Hex(),
		}).Info("deploy: Migrated deployment")

		return
	}()
}

func (d *Deployments) destroy(db *database.Database,
	deply *deployment.Deployment) {

	acquired, lockId := deploymentsLock.LockOpenTimeout(
		deply.Id.Hex(), 3*time.Minute)
	if !acquired {
		return
	}

	go func() {
		defer func() {
			deploymentsLock.Unlock(deply.Id.Hex(), lockId)
		}()

		if deply.Node != d.stat.Node().Id {
			return
		}

		if deply.Kind == deployment.Image && !deply.Image.IsZero() {
			img, err := image.Get(db, deply.Image)
			if err != nil {
				if _, ok := err.(*database.NotFoundError); ok {
					img = nil
					err = nil
				} else {
					logrus.WithFields(logrus.Fields{
						"deployment_id": deply.Id.Hex(),
						"error":         err,
					}).Error("deploy: Failed to get image")
					return
				}
			}

			if img != nil {
				err = img.Remove(db)
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"deployment_id": deply.Id.Hex(),
						"error":         err,
					}).Error("deploy: Failed to remove deployment image")
					return
				}

				event.PublishDispatch(db, "image.change")
			}
		}

		inst := d.stat.GetInstace(deply.Instance)
		if inst != nil {
			if inst.DeleteProtection {
				logrus.WithFields(logrus.Fields{
					"deployment": deply.Id.Hex(),
					"instance":   inst.Id.Hex(),
				}).Warning("deploy: Cannot destroy deployment with " +
					"instance delete protection")
				return
			}

			if inst.State != instance.Destroy {
				logrus.WithFields(logrus.Fields{
					"deployment": deply.Id.Hex(),
					"instance":   inst.Id.Hex(),
				}).Info("deploy: Delete deployment instance")

				err := instance.Delete(db, inst.Id)
				if err != nil {
					if _, ok := err.(*database.NotFoundError); !ok {
						err = nil
					} else {
						logrus.WithFields(logrus.Fields{
							"deployment_id": deply.Id.Hex(),
							"error":         err,
						}).Error("deploy: Failed to delete instance")
						return
					}
				}
			}
		} else {
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

		return
	}()
}

func (d *Deployments) archive(db *database.Database,
	deply *deployment.Deployment) (err error) {

	if deply.Node != d.stat.Node().Id {
		return
	}

	inst := d.stat.GetInstace(deply.Instance)
	if inst == nil {
		return
	}

	if inst.State != instance.Stop {
		logrus.WithFields(logrus.Fields{
			"instance_id": inst.Id.Hex(),
		}).Info("deploy: Stopping instance for deployment archive")

		err = instance.SetState(db, inst.Id, instance.Stop)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"instance_id": inst.Id.Hex(),
				"error":       err,
			}).Error("deploy: Failed to set instance state")

			return
		}
	}

	return
}

func (d *Deployments) restore(db *database.Database,
	deply *deployment.Deployment) (err error) {

	if deply.Node != d.stat.Node().Id {
		return
	}

	inst := d.stat.GetInstace(deply.Instance)
	if inst == nil {
		return
	}

	if inst.State != instance.Start {
		logrus.WithFields(logrus.Fields{
			"instance_id": inst.Id.Hex(),
		}).Info("deploy: Starting instance for deployment restore")

		err = instance.SetState(db, inst.Id, instance.Start)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"instance_id": inst.Id.Hex(),
				"error":       err,
			}).Error("deploy: Failed to set instance state")

			return
		}
	}

	return
}

func (d *Deployments) imageShutdown(deply *deployment.Deployment,
	virt *vm.VirtualMachine) {

	acquired, lockId := deploymentsLock.LockOpenTimeout(
		deply.Id.Hex(), 5*time.Minute)
	if !acquired {
		return
	}

	go func() {
		defer func() {
			time.Sleep(3 * time.Second)
			deploymentsLock.Unlock(deply.Id.Hex(), lockId)
		}()

		db := database.GetDatabase()
		defer db.Close()

		logrus.WithFields(logrus.Fields{
			"instance_id": virt.Id.Hex(),
		}).Info("deploy: Stopping instance for deployment image")

		time.Sleep(3 * time.Second)

		err := imds.Pull(db, virt.Id, virt.Deployment, virt.ImdsHostSecret)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"instance_id": virt.Id.Hex(),
				"error":       err,
			}).Error("deploy: Failed to pull imds state for shutdown")
		}

		err = instance.SetState(db, virt.Id, instance.Stop)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"instance_id": virt.Id.Hex(),
				"error":       err,
			}).Error("deploy: Failed to set instance state")

			return
		}

		if deply.GetImageState() == "" {
			deply.SetImageState(deployment.Ready)
			err = deply.CommitFields(db, set.NewSet("image_data.state"))
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"instance_id": virt.Id.Hex(),
					"error":       err,
				}).Error("deploy: Failed to commit deployment state")
				return
			}
		}

		event.PublishDispatch(db, "pod.change")
	}()
}

func (d *Deployments) image(db *database.Database,
	deply *deployment.Deployment) (err error) {

	if deply.Node != d.stat.Node().Id {
		return
	}

	inst := d.stat.GetInstace(deply.Instance)
	if inst == nil {
		return
	}

	virt := d.stat.GetVirt(inst.Id)

	if inst.Guest == nil {
		return
	}

	if inst.IsActive() && inst.Guest.Status == types.Imaged &&
		inst.State != instance.Stop {

		d.imageShutdown(deply, virt)
		return
	}

	if deply.State == deployment.Deployed &&
		deply.GetImageState() == deployment.Ready &&
		!inst.IsActive() && inst.Guest.Status == types.Imaged {

		logrus.WithFields(logrus.Fields{
			"instance_id": inst.Id.Hex(),
		}).Info("deploy: Creating deployment image")

		dsk, e := disk.GetInstanceIndex(db, inst.Id, "0")
		if e != nil {
			if _, ok := e.(*database.NotFoundError); ok {
				logrus.WithFields(logrus.Fields{
					"instance_id": inst.Id.Hex(),
					"error":       err,
				}).Error("deploy: Failed to find instance disk for image")

				deply.SetImageState(deployment.Failed)
				err = deply.CommitFields(db, set.NewSet("image_data.state"))
				if err != nil {
					return
				}

				dsk = nil
				err = nil
			} else {
				return
			}
		}

		dsk.State = disk.Snapshot
		err = dsk.CommitFields(db, set.NewSet("state"))
		if err != nil {
			return
		}

		deply.SetImageState(deployment.Snapshot)
		err = deply.CommitFields(db, set.NewSet("image_data.state"))
		if err != nil {
			return
		}
	}

	return
}

func (d *Deployments) domainCommit(deply *deployment.Deployment,
	domn *domain.Domain, newRecs []*domain.Record) {

	acquired, lockId := deploymentsLock.LockOpenTimeout(
		deply.Id.Hex(), 3*time.Minute)
	if !acquired {
		return
	}

	go func() {
		defer func() {
			deploymentsLock.Unlock(deply.Id.Hex(), lockId)
		}()

		db := database.GetDatabase()
		defer db.Close()

		logrus.WithFields(logrus.Fields{
			"domain_id": domn.Id.Hex(),
		}).Info("deploy: Committing domain records")

		recs := []*deployment.RecordData{}
		for _, rec := range newRecs {
			recs = append(recs, &deployment.RecordData{
				Domain: rec.SubDomain + "." + domn.RootDomain,
				Value:  rec.Value,
			})
		}
		domnData := &deployment.DomainData{
			Records: recs,
		}

		err := domn.CommitRecords(db)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"domain_id": domn.Id.Hex(),
				"error":     err,
			}).Error("deploy: Failed to commit domain records")
			return
		}

		deply.DomainData = domnData
		err = deply.CommitFields(db, set.NewSet("domain_data"))
		if err != nil {
			return
		}

		event.PublishDispatch(db, "domain.change")
	}()
}

func (d *Deployments) domain(db *database.Database,
	deply *deployment.Deployment, spc *spec.Spec) (err error) {

	if spc.Domain != nil && deply.InstanceData != nil {
		newRecs := map[primitive.ObjectID][]*domain.Record{}

		for _, specRec := range spc.Domain.Records {
			domn := d.stat.SpecDomain(specRec.Domain)
			if domn == nil {
				continue
			}

			switch specRec.Type {
			case spec.Public:
				for _, val := range deply.InstanceData.PublicIps {
					rec := &domain.Record{
						Domain:     specRec.Domain,
						SubDomain:  specRec.Name,
						Deployment: deply.Id,
						Type:       domain.A,
						Value:      val,
					}

					errData, e := rec.Validate(db)
					if e != nil {
						err = e
						return
					}
					if errData != nil {
						err = errData.GetError()
						return
					}

					newRecs[domn.Id] = append(newRecs[domn.Id], rec)
				}
				break
			case spec.Public6:
				for _, val := range deply.InstanceData.PublicIps6 {
					rec := &domain.Record{
						Domain:     specRec.Domain,
						SubDomain:  specRec.Name,
						Deployment: deply.Id,
						Type:       domain.AAAA,
						Value:      val,
					}

					errData, e := rec.Validate(db)
					if e != nil {
						err = e
						return
					}
					if errData != nil {
						err = errData.GetError()
						return
					}

					newRecs[domn.Id] = append(newRecs[domn.Id], rec)
				}
				break
			case spec.Private:
				for _, val := range deply.InstanceData.PrivateIps {
					rec := &domain.Record{
						Domain:     specRec.Domain,
						SubDomain:  specRec.Name,
						Deployment: deply.Id,
						Type:       domain.A,
						Value:      val,
					}

					errData, e := rec.Validate(db)
					if e != nil {
						err = e
						return
					}
					if errData != nil {
						err = errData.GetError()
						return
					}

					newRecs[domn.Id] = append(newRecs[domn.Id], rec)
				}
				break
			case spec.Private6:
				for _, val := range deply.InstanceData.PrivateIps6 {
					rec := &domain.Record{
						Domain:     specRec.Domain,
						SubDomain:  specRec.Name,
						Deployment: deply.Id,
						Type:       domain.AAAA,
						Value:      val,
					}

					errData, e := rec.Validate(db)
					if e != nil {
						err = e
						return
					}
					if errData != nil {
						err = errData.GetError()
						return
					}

					newRecs[domn.Id] = append(newRecs[domn.Id], rec)
				}
				break
			case spec.OraclePublic:
				for _, val := range deply.InstanceData.OraclePublicIps {
					rec := &domain.Record{
						Domain:     specRec.Domain,
						SubDomain:  specRec.Name,
						Deployment: deply.Id,
						Type:       domain.A,
						Value:      val,
					}

					errData, e := rec.Validate(db)
					if e != nil {
						err = e
						return
					}
					if errData != nil {
						err = errData.GetError()
						return
					}

					newRecs[domn.Id] = append(newRecs[domn.Id], rec)
				}
				break
			case spec.OraclePrivate:
				for _, val := range deply.InstanceData.OraclePrivateIps {
					rec := &domain.Record{
						Domain:     specRec.Domain,
						SubDomain:  specRec.Name,
						Deployment: deply.Id,
						Type:       domain.A,
						Value:      val,
					}

					errData, e := rec.Validate(db)
					if e != nil {
						err = e
						return
					}
					if errData != nil {
						err = errData.GetError()
						return
					}

					newRecs[domn.Id] = append(newRecs[domn.Id], rec)
				}
				break
			}
		}

		for domnId, domnNewRecs := range newRecs {
			domn := d.stat.SpecDomain(domnId)
			if domn == nil {
				continue
			}

			changedDomn := domn.MergeRecords(deply.Id, domnNewRecs)
			if changedDomn != nil {
				d.domainCommit(deply, changedDomn, domnNewRecs)
			}
		}
	}

	return
}

func (d *Deployments) Deploy(db *database.Database) (err error) {
	activeDeployments := d.stat.DeploymentsDeployed()
	inactiveDeployments := d.stat.DeploymentsInactive()

	for _, deply := range inactiveDeployments {
		switch deply.State {
		case deployment.Destroy:
			d.destroy(db, deply)
			break
		case deployment.Archive:
			err = d.archive(db, deply)
			if err != nil {
				return
			}
			break
		case deployment.Restore:
			err = d.restore(db, deply)
			if err != nil {
				return
			}
			break
		}
	}

	for _, deply := range activeDeployments {
		if deply.State == deployment.Migrate {
			d.migrate(db, deply)
			break
		}

		if deply.State == deployment.Deployed &&
			deply.Kind == deployment.Instance {

			spec := d.stat.Spec(deply.Spec)
			if spec != nil && spec.Domain != nil {
				err = d.domain(db, deply, spec)
				if err != nil {
					return
				}
			}
		}

		if deply.Kind == deployment.Image {
			err = d.image(db, deply)
			if err != nil {
				return
			}
		}

	}

	return
}

func NewDeployments(stat *state.State) *Deployments {
	return &Deployments{
		stat: stat,
	}
}

package deploy

import (
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/deployment"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/image"
	"github.com/pritunl/pritunl-cloud/imds"
	"github.com/pritunl/pritunl-cloud/imds/types"
	"github.com/pritunl/pritunl-cloud/instance"
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

func (d *Deployments) destroy(db *database.Database,
	deply *deployment.Deployment) (err error) {

	if deply.Node != d.stat.Node().Id {
		return
	}

	if deply.Kind == deployment.Image && !deply.Image.IsZero() {
		img, e := image.Get(db, deply.Image)
		if e != nil {
			err = e
			if _, ok := err.(*database.NotFoundError); ok {
				img = nil
				err = nil
			} else {
				return
			}

			return
		}

		if img != nil {
			err = img.Remove(db)
			if err != nil {
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

			time.Sleep(5 * time.Second)

			err = instance.Delete(db, inst.Id)
			if err != nil {
				if _, ok := err.(*database.NotFoundError); !ok {
					err = nil
				} else {
					return
				}
			}
		}
	} else {
		err = deployment.Remove(db, deply.Id)
		if err != nil {
			return
		}
	}

	return
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
		virt.Id.Hex(), 10*time.Minute)
	if !acquired {
		return
	}

	go func() {
		defer func() {
			time.Sleep(3 * time.Second)
			deploymentsLock.Unlock(virt.Id.Hex(), lockId)
		}()

		db := database.GetDatabase()
		defer db.Close()

		logrus.WithFields(logrus.Fields{
			"instance_id": virt.Id.Hex(),
		}).Info("deploy: Stopping instance for deployment image")

		time.Sleep(3 * time.Second)

		err := imds.Pull(db, virt.Id, virt.ImdsHostSecret)
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

		event.PublishDispatch(db, "service.change")
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

func (d *Deployments) Deploy(db *database.Database) (err error) {
	activeDeployments := d.stat.DeploymentsDeployed()
	inactiveDeployments := d.stat.DeploymentsInactive()

	for _, deply := range inactiveDeployments {
		switch deply.State {
		case deployment.Destroy:
			err = d.destroy(db, deply)
			if err != nil {
				return
			}
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
		if deply.Kind != deployment.Image {
			continue
		}

		err = d.image(db, deply)
		if err != nil {
			return
		}
	}

	return
}

func NewDeployments(stat *state.State) *Deployments {
	return &Deployments{
		stat: stat,
	}
}

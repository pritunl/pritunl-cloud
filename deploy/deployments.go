package deploy

import (
	"time"

	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/state"
	"github.com/sirupsen/logrus"
)

type Deployments struct {
	stat *state.State
}

func (d *Deployments) Deploy(db *database.Database) (err error) {
	destroyDeployments := d.stat.DeploymentsDestroy()

	for _, deply := range destroyDeployments {
		if deply.Node != d.stat.Node().Id {
			continue
		}

		inst := d.stat.GetInstace(deply.Instance)
		if inst == nil {
			continue
		}

		if inst.DeleteProtection {
			logrus.WithFields(logrus.Fields{
				"deployment": deply.Id.Hex(),
				"instance":   inst.Id.Hex(),
			}).Warning("deploy: Cannot destroy deployment with " +
				"instance delete protection")
			continue
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
	}

	return
}

func NewDeployments(stat *state.State) *Deployments {
	return &Deployments{
		stat: stat,
	}
}

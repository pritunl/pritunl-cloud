package deploy

import (
	"github.com/pritunl/pritunl-cloud/hnetwork"
	"github.com/pritunl/pritunl-cloud/interfaces"
	"github.com/pritunl/pritunl-cloud/networking"
	"github.com/pritunl/pritunl-cloud/state"
	"github.com/pritunl/pritunl-cloud/vxlan"
	"github.com/sirupsen/logrus"
)

type Network struct {
	stat *state.State
}

func (d *Network) Deploy() (err error) {
	err = networking.ApplyState(d.stat)
	if err != nil {
		return
	}

	err = hnetwork.ApplyState(d.stat)
	if err != nil {
		return
	}

	err = vxlan.ApplyState(d.stat)
	if err != nil {
		return
	}

	err = ApplyOracleState(d.stat)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("deploy: Failed to apply Oracle state")
		err = nil
	}

	interfaces.SyncIfaces(d.stat.VxLan())

	return
}

func NewNetwork(stat *state.State) *Network {
	return &Network{
		stat: stat,
	}
}

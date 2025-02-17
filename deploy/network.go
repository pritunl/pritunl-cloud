package deploy

import (
	"github.com/pritunl/pritunl-cloud/hnetwork"
	"github.com/pritunl/pritunl-cloud/interfaces"
	"github.com/pritunl/pritunl-cloud/state"
	"github.com/pritunl/pritunl-cloud/vxlan"
)

type Network struct {
	stat *state.State
}

func (d *Network) Deploy() (err error) {
	err = hnetwork.ApplyState(d.stat)
	if err != nil {
		return
	}

	err = vxlan.ApplyState(d.stat)
	if err != nil {
		return
	}

	interfaces.SyncIfaces(d.stat.VxLan())

	return
}

func NewNetwork(stat *state.State) *Network {
	return &Network{
		stat: stat,
	}
}

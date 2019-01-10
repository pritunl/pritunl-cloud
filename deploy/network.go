package deploy

import (
	"github.com/pritunl/pritunl-cloud/networking"
	"github.com/pritunl/pritunl-cloud/state"
)

type Network struct {
	stat *state.State
}

func (d *Network) Deploy() (err error) {
	err = networking.ApplyState(d.stat)
	if err != nil {
		return
	}

	return
}

func NewNetwork(stat *state.State) *Network {
	return &Network{
		stat: stat,
	}
}

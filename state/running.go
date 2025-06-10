package state

import (
	"io/ioutil"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

var (
	Running    = &RunningState{}
	RunningPkg = NewPackage(Running)
)

type RunningState struct {
	running []string
}

func (p *RunningState) Running() []string {
	return p.running
}

func (p *RunningState) Refresh(pkg *Package,
	db *database.Database) (err error) {

	items, err := ioutil.ReadDir("/var/run")
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "state: Failed to read run directory"),
		}
		return
	}

	running := []string{}
	for _, item := range items {
		if !item.IsDir() {
			running = append(running, item.Name())
		}
	}
	p.running = running

	return
}

func (p *RunningState) Apply(st *State) {
	st.Running = p.Running
}

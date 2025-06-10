package state

import (
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/scheduler"
)

var (
	Schedulers    = &SchedulersState{}
	SchedulersPkg = NewPackage(Schedulers)
)

type SchedulersState struct {
	schedulers []*scheduler.Scheduler
}

func (p *SchedulersState) Schedulers() []*scheduler.Scheduler {
	return p.schedulers
}

func (p *SchedulersState) Refresh(pkg *Package,
	db *database.Database) (err error) {

	schedulers, err := scheduler.GetAll(db)
	if err != nil {
		return
	}
	p.schedulers = schedulers

	return
}

func (p *SchedulersState) Apply(st *State) {
	st.Schedulers = p.Schedulers
}

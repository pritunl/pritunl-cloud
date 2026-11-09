package statec

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/scheduler"
)

var (
	Schedulers      = &SchedulersState{}
	SchedulersPkg   = NewPackage(Schedulers)
	schedulersCache = NewCache(
		"schedulers",
		nil,
		func() Scope {
			return Scope{
				Filter: bson.M{},
				Key:    "all",
				Ok:     true,
			}
		},
		nil,
	)
)

type SchedulersState struct {
	schedulers []*scheduler.Scheduler
}

func (p *SchedulersState) Schedulers() []*scheduler.Scheduler {
	return p.schedulers
}

func (p *SchedulersState) Refresh(pkg *Package,
	db *database.Database) (err error) {

	err = schedulersCache.Sync(db)
	if err != nil {
		return
	}

	schedulers, err := Decode[scheduler.Scheduler](schedulersCache)
	if err != nil {
		return
	}
	p.schedulers = schedulers

	return
}

func (p *SchedulersState) Apply(st *State) {
	st.Schedulers = p.Schedulers
}

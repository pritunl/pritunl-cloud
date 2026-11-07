package statec

import (
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/state"
)

type State = state.State

type Runtimes = state.Runtimes

func GetState(runtimes *Runtimes) (stat *State, err error) {
	Watch.Ensure()

	db := database.GetDatabase()
	defer db.Close()

	err = RefreshAll(db, runtimes)
	if err != nil {
		return
	}

	stat = state.NewState()
	ApplyAll(stat)

	return
}

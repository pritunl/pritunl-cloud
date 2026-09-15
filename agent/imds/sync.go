package imds

import (
	"sync"

	"github.com/pritunl/pritunl-cloud/imds/types"
	"github.com/pritunl/pritunl-cloud/telemetry"
)

var (
	curStatus     = types.Initializing
	curStatusLock sync.Mutex
)

type StateData struct {
	*types.State
}

func (m *Imds) GetState(curHash uint32) (data *StateData, err error) {
	data = &StateData{
		&types.State{},
	}

	data.Hash = curHash
	curStatusLock.Lock()
	data.Status = curStatus
	curStatusLock.Unlock()

	data.Metrics = telemetry.Metrics.GetAll()

	data.Updates, data.Components = telemetry.GetUpdates()

	return
}

func SetStatus(status string) {
	curStatusLock.Lock()
	curStatus = status
	curStatusLock.Unlock()
}

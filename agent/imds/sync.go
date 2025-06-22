package imds

import (
	"sync"
	"time"

	"github.com/pritunl/pritunl-cloud/imds/types"
	"github.com/pritunl/pritunl-cloud/telemetry"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/tools/logger"
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

	mem, err := utils.GetMemInfo()
	if err != nil {
		logger.WithFields(logger.Fields{
			"error": err,
		}).Limit(30 * time.Minute).Error("imds: Failed to get memory")
	} else {
		data.Memory = utils.ToFixed(mem.UsedPercent, 2)
		data.HugePages = utils.ToFixed(mem.HugePagesUsedPercent, 2)
	}

	load, err := utils.LoadAverage()
	if err != nil {
		logger.WithFields(logger.Fields{
			"error": err,
		}).Limit(30 * time.Minute).Error("imds: Failed to get load")
	} else {
		data.Load1 = load.Load1
		data.Load5 = load.Load5
		data.Load15 = load.Load15
	}

	data.Updates = telemetry.Updates.Get()

	return
}

func SetStatus(status string) {
	curStatusLock.Lock()
	curStatus = status
	curStatusLock.Unlock()
}

package imds

import (
	"github.com/pritunl/pritunl-cloud/imds/types"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/tools/logger"
)

var (
	curStatus = types.Initializing
)

type StateData struct {
	*types.State
}

func (m *Imds) GetState() (data *StateData, err error) {
	data = &StateData{
		&types.State{},
	}

	data.Status = curStatus

	mem, err := utils.GetMemInfo()
	if err != nil {
		logger.WithFields(logger.Fields{
			"error": err,
		}).Error("imds: Failed to get memory")
	} else {
		data.Memory = utils.ToFixed(mem.UsedPercent, 2)
		data.HugePages = utils.ToFixed(mem.HugePagesUsedPercent, 2)
	}

	load, err := utils.LoadAverage()
	if err != nil {
		logger.WithFields(logger.Fields{
			"error": err,
		}).Error("imds: Failed to get load")
	} else {
		data.Load1 = load.Load1
		data.Load5 = load.Load5
		data.Load15 = load.Load15
	}

	return
}

func SetStatus(status string) {
	curStatus = status
}

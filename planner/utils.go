package planner

import (
	"time"

	"github.com/pritunl/pritunl-cloud/eval"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/plan"
	"github.com/pritunl/pritunl-cloud/unit"
	"github.com/pritunl/pritunl-cloud/utils"
)

func buildEvalData(unt *unit.Unit,
	inst *instance.Instance) (data eval.Data, err error) {

	lastTimestamp := 0
	lastHeartbeat := 0
	if inst.IsActive() {
		now := time.Now()

		uptime := int(now.Sub(inst.VirtTimestamp).Seconds())
		if inst.Guest != nil {
			lastTimestamp = int(now.Sub(inst.Guest.Timestamp).Seconds())
			lastHeartbeat = int(now.Sub(inst.Guest.Heartbeat).Seconds())
		}
		lastTimestamp = utils.Min(lastTimestamp, uptime)
		lastHeartbeat = utils.Min(lastHeartbeat, uptime)
	}

	dataStrct := plan.Data{
		Unit: plan.Unit{
			Name:  unt.Name,
			Count: unt.Count,
		},
		Instance: plan.Instance{
			Name:          inst.Name,
			State:         inst.State,
			Action:        inst.Action,
			VirtState:     inst.VirtState,
			LastTimestamp: lastTimestamp,
			LastHeartbeat: lastHeartbeat,
		},
	}

	data, err = dataStrct.Export()
	if err != nil {
		return
	}

	return
}

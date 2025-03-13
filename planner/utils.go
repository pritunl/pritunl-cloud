package planner

import (
	"time"

	"github.com/pritunl/pritunl-cloud/eval"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/plan"
	"github.com/pritunl/pritunl-cloud/pod"
	"github.com/pritunl/pritunl-cloud/utils"
)

func buildEvalData(servc *pod.Pod, unit *pod.Unit,
	inst *instance.Instance) (data eval.Data, err error) {

	lastHeartbeat := 0
	if inst.IsActive() {
		now := time.Now()

		uptime := int(now.Sub(inst.VirtTimestamp).Seconds())
		if inst.Guest != nil {
			lastHeartbeat = int(now.Sub(inst.Guest.Heartbeat).Seconds())
		}
		lastHeartbeat = utils.Min(lastHeartbeat, uptime)
	}

	dataStrct := plan.Data{
		Pod: plan.Pod{
			Name: servc.Name,
		},
		Unit: plan.Unit{
			Name:  unit.Name,
			Count: unit.Count,
		},
		Instance: plan.Instance{
			Name:          inst.Name,
			State:         inst.State,
			Action:        inst.Action,
			VirtState:     inst.VirtState,
			LastHeartbeat: lastHeartbeat,
		},
	}

	data, err = dataStrct.Export()
	if err != nil {
		return
	}

	return
}

package planner

import (
	"github.com/pritunl/pritunl-cloud/eval"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/plan"
	"github.com/pritunl/pritunl-cloud/pod"
)

func buildEvalData(servc *pod.Pod, unit *pod.Unit,
	inst *instance.Instance) (data eval.Data, err error) {

	dataStrct := plan.Data{
		Pod: plan.Pod{
			Name: servc.Name,
		},
		Unit: plan.Unit{
			Name:  unit.Name,
			Count: unit.Count,
		},
		Instance: plan.Instance{
			Name:      inst.Name,
			State:     inst.State,
			VirtState: inst.VirtState,
		},
	}

	data, err = dataStrct.Export()
	if err != nil {
		return
	}

	return
}

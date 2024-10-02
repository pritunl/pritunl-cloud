package planner

import (
	"github.com/pritunl/pritunl-cloud/eval"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/plan"
	"github.com/pritunl/pritunl-cloud/service"
)

func buildEvalData(servc *service.Service, unit *service.Unit,
	inst *instance.Instance) (data eval.Data, err error) {

	dataStrct := plan.Data{
		Service: plan.Service{
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

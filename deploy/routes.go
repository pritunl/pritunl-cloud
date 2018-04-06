package deploy

import (
	"fmt"
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/state"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/pritunl/pritunl-cloud/vpc"
	"strings"
)

type Routes struct {
	stat *state.State
}

func (n *Routes) update(inst *instance.Instance) (err error) {
	vc := n.stat.Vpc(inst.Vpc)
	if vc == nil {
		err = &errortypes.NotFoundError{
			errors.New("deploy: Instance vpc not found"),
		}
		return
	}

	namespace := vm.GetNamespace(inst.Id, 0)

	curRoutes := set.NewSet()
	newRoutes := set.NewSet()

	output, err := utils.ExecCombinedOutput(
		"ip", "netns", "exec", namespace,
		"route", "-n",
	)
	if err != nil {
		err = nil
		return
	}

	lines := strings.Split(output, "\n")
	if len(lines) > 2 {
		for _, line := range lines[2:] {
			if line == "" {
				continue
			}

			fields := strings.Fields(line)
			if len(fields) < 8 {
				continue
			}

			if fields[4] != "97" {
				continue
			}

			mask := utils.ParseIpMask(fields[2])
			if mask == nil {
				continue
			}
			cidr, _ := mask.Size()

			route := vpc.Route{
				Destination: fmt.Sprintf("%s/%d", fields[0], cidr),
				Target:      fields[1],
			}

			curRoutes.Add(route)

		}
	}

	if vc.Routes != nil {
		for _, route := range vc.Routes {
			newRoutes.Add(*route)
		}
	}

	addRoutes := newRoutes.Copy()
	remRoutes := curRoutes.Copy()

	addRoutes.Subtract(curRoutes)
	remRoutes.Subtract(newRoutes)

	for routeInf := range remRoutes.Iter() {
		route := routeInf.(vpc.Route)

		utils.ExecCombinedOutputLogged(
			nil,
			"ip", "netns", "exec", namespace,
			"ip", "route",
			"del", route.Destination,
			"via", route.Target,
			"metric", "97",
		)
	}

	for routeInf := range addRoutes.Iter() {
		route := routeInf.(vpc.Route)

		utils.ExecCombinedOutputLogged(
			nil,
			"ip", "netns", "exec", namespace,
			"ip", "route",
			"add", route.Destination,
			"via", route.Target,
			"metric", "97",
		)
	}

	return
}

func (n *Routes) Deploy() (err error) {
	instances := n.stat.Instances()

	for _, inst := range instances {
		err = n.update(inst)
		if err != nil {
			return
		}
	}

	return
}

func NewRoutes(stat *state.State) *Routes {
	return &Routes{
		stat: stat,
	}
}

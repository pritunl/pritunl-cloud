package deploy

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/pritunl-cloud/state"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"strings"
)

type Namespace struct {
	stat *state.State
}

func (n *Namespace) Deploy() (err error) {
	instances := n.stat.Instances()

	curNamespaces := set.NewSet()
	curVirtIfaces := set.NewSet()

	for _, inst := range instances {
		curNamespaces.Add(vm.GetNamespace(inst.Id, 0))
		curVirtIfaces.Add(vm.GetIfaceVirt(inst.Id, 0))
	}

	output, err := utils.ExecOutputLogged(
		nil, "ip", "-o", "link", "show",
	)
	if err != nil {
		return
	}

	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 || len(fields[1]) < 2 {
			continue
		}
		iface := strings.Split(fields[1][:len(fields[1])-1], "@")[0]

		if len(iface) != 14 || !strings.HasPrefix(iface, "v") {
			continue
		}

		if !curVirtIfaces.Contains(iface) {
			utils.ExecCombinedOutputLogged(
				nil,
				"ip", "link", "del", iface,
			)
		}
	}

	output, err = utils.ExecOutputLogged(
		nil,
		"ip", "netns", "list",
	)
	if err != nil {
		return
	}

	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		namespace := fields[0]

		if !curNamespaces.Contains(namespace) {
			_, err = utils.ExecCombinedOutputLogged(
				nil,
				"ip", "netns", "del", namespace,
			)
			if err != nil {
				return
			}
		}
	}

	return
}

func NewNamespace(stat *state.State) *Namespace {
	return &Namespace{
		stat: stat,
	}
}

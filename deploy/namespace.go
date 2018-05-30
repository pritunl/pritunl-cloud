package deploy

import (
	"fmt"
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/state"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Namespace struct {
	stat *state.State
}

func (n *Namespace) Deploy() (err error) {
	instances := n.stat.Instances()

	curNamespaces := set.NewSet()
	curVirtIfaces := set.NewSet()
	curInternalIfaces := set.NewSet()

	for _, inst := range instances {
		curNamespaces.Add(vm.GetNamespace(inst.Id, 0))
		curVirtIfaces.Add(vm.GetIfaceVirt(inst.Id, 0))
		curInternalIfaces.Add(vm.GetIfaceInternal(inst.Id, 0))
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
		if len(namespace) != 14 || !strings.HasPrefix(namespace, "n") {
			continue
		}

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

	items, err := ioutil.ReadDir("/var/run")
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "deploy: Failed to read run directory"),
		}
		return
	}

	for _, item := range items {
		name := item.Name()

		if item.IsDir() || len(name) != 27 ||
			!strings.HasPrefix(name, "dhclient-i") {

			continue
		}

		iface := name[9:23]

		if !curInternalIfaces.Contains(iface) {
			pth := filepath.Join("/var/run", item.Name())

			pidByt, e := ioutil.ReadFile(pth)
			if e != nil {
				err = &errortypes.ReadError{
					errors.Wrap(e, "ipsec: Failed to read dhclient pid"),
				}
				return
			}

			pid, e := strconv.Atoi(strings.TrimSpace(string(pidByt)))
			if e != nil {
				err = &errortypes.ParseError{
					errors.Wrap(e, "ipsec: Failed to parse dhclient pid"),
				}
				return
			}

			exists, _ := utils.Exists(fmt.Sprintf("/proc/%d/status", pid))
			if exists {
				utils.ExecCombinedOutput("", "kill", "-9", strconv.Itoa(pid))
			} else {
				os.Remove(pth)
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

package deploy

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/interfaces"
	"github.com/pritunl/pritunl-cloud/ipsec"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/state"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/pritunl/pritunl-cloud/vpc"
)

type Ipsec struct {
	stat *state.State
}

func (t *Ipsec) Deploy() (err error) {
	db := database.GetDatabase()
	defer db.Close()

	curVpcs := set.NewSet()
	curNamespaces := set.NewSet()
	curVirtIfaces := set.NewSet()
	curExternalIfaces := set.NewSet()
	newState := []*vpc.Vpc{}

	vpcs := t.stat.Vpcs()
	for _, vc := range vpcs {
		if vc.LinkUris == nil || len(vc.LinkUris) == 0 {
			continue
		}

		if vc.LinkNode != node.Self.Id &&
			time.Since(vc.LinkTimestamp) < time.Duration(
				settings.Ipsec.LinkTimeout)*time.Second {

			continue
		}

		curVpcs.Add(vc.Id)
		curNamespaces.Add(vm.GetLinkNamespace(vc.Id, 0))
		curVirtIfaces.Add(vm.GetLinkIfaceVirt(vc.Id, 0))
		curVirtIfaces.Add(vm.GetLinkIfaceVirt(vc.Id, 1))
		curExternalIfaces.Add(vm.GetLinkIfaceExternal(vc.Id, 0))

		newState = append(newState, vc)
	}

	ipsec.ApplyState(newState)

	ifaces := t.stat.Interfaces()
	for _, iface := range ifaces {
		if len(iface) != 14 || !strings.HasPrefix(iface, "y") {
			continue
		}

		if !curVirtIfaces.Contains(iface) {
			utils.ExecCombinedOutputLogged(
				nil,
				"ip", "link", "del", iface,
			)
			interfaces.RemoveVirtIface(iface)
		}
	}

	namespaces := t.stat.Namespaces()
	for _, namespace := range namespaces {
		if len(namespace) != 14 || !strings.HasPrefix(namespace, "x") {
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

	running := t.stat.Running()
	for _, name := range running {
		if len(name) != 27 || !strings.HasPrefix(name, "dhclient-z") {
			continue
		}

		iface := name[9:23]

		if !curExternalIfaces.Contains(iface) {
			pth := filepath.Join("/var/run", name)

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

func NewIpsec(stat *state.State) *Ipsec {
	return &Ipsec{
		stat: stat,
	}
}

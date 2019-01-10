package networking

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/pritunl-cloud/interfaces"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/state"
	"github.com/pritunl/pritunl-cloud/utils"
	"strconv"
)

var (
	curIfaces = set.NewSet()
	curMtu    = 0
)

func ApplyState(stat *state.State) (err error) {
	newMtu := 0

	if node.Self.JumboFrames {
		newMtu = settings.Hypervisor.JumboMtu
	} else {
		newMtu = settings.Hypervisor.NormalMtu
	}

	if newMtu != curMtu {
		curIfaces = set.NewSet()
		curMtu = newMtu
	}

	mtu := strconv.Itoa(newMtu)
	bridges := interfaces.GetBridges()

	ifaces := stat.Interfaces()
	for _, iface := range ifaces {
		if len(iface) == 14 || bridges.Contains(iface) ||
			curIfaces.Contains(iface) {

			continue
		}

		upper, e := utils.GetInterfaceUpper(iface)
		if e != nil {
			err = e
			return
		}

		if upper == "" {
			continue
		}

		if bridges.Contains(upper) {
			_, e = utils.ExecCombinedOutputLogged(
				nil,
				"ip", "link",
				"set", "dev", iface,
				"mtu", mtu,
			)
			if e != nil {
				continue
			}

			curIfaces.Add(iface)
		}
	}

	return
}

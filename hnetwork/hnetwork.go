package hnetwork

import (
	"github.com/sirupsen/logrus"
	"github.com/pritunl/pritunl-cloud/iproute"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/state"
	"github.com/pritunl/pritunl-cloud/utils"
)

var (
	initialized = false
	curState    = ""
)

func removeNetwork(stat *state.State) (err error) {
	if curState != "" || stat.HasInterfaces(
		settings.Hypervisor.HostNetworkName) {

		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "link", "set",
			"dev", settings.Hypervisor.HostNetworkName, "down",
		)
		if err != nil {
			return
		}

		_ = iproute.BridgeDelete("", settings.Hypervisor.HostNetworkName)

		curState = ""
	}

	return
}

func ApplyState(stat *state.State) (err error) {
	if !initialized {
		addr, e := getAddr()
		if e != nil {
			err = e
			return
		}

		initialized = true
		curState = addr
	}

	hostBlock := stat.NodeHostBlock()
	if stat.NodeHostBlock() != nil {
		gatewayCidr := hostBlock.GetGatewayCidr()
		if gatewayCidr == "" {
			logrus.WithFields(logrus.Fields{
				"host_block": hostBlock.Id.Hex(),
			}).Error("hnetwork: Host network block gateway is invalid")

			err = removeNetwork(stat)
			if err != nil {
				return
			}

			return
		}

		if curState != gatewayCidr {
			logrus.WithFields(logrus.Fields{
				"host_block":         hostBlock.Id.Hex(),
				"host_block_gateway": gatewayCidr,
			}).Info("hnetwork: Updating host network bridge")

			err = iproute.BridgeAdd("", settings.Hypervisor.HostNetworkName)
			if err != nil {
				return
			}

			err = setAddr(gatewayCidr)
			if err != nil {
				return
			}

			curState = gatewayCidr
		}
	} else {
		err = removeNetwork(stat)
		if err != nil {
			return
		}
	}

	return
}

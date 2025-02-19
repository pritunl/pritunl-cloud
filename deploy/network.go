package deploy

import (
	"fmt"

	"github.com/pritunl/pritunl-cloud/block"
	"github.com/pritunl/pritunl-cloud/bridges"
	"github.com/pritunl/pritunl-cloud/hnetwork"
	"github.com/pritunl/pritunl-cloud/interfaces"
	"github.com/pritunl/pritunl-cloud/iproute"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/state"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vxlan"
	"github.com/sirupsen/logrus"
)

var (
	nodePortInitialized = false
	nodePortCurGateway  = ""
)

type Network struct {
	stat *state.State
}

func (d *Network) Deploy() (err error) {
	err = hnetwork.ApplyState(d.stat)
	if err != nil {
		return
	}

	err = NodePortApplyState(d.stat)
	if err != nil {
		return
	}

	err = vxlan.ApplyState(d.stat)
	if err != nil {
		return
	}

	interfaces.SyncIfaces(d.stat.VxLan())

	return
}

func NewNetwork(stat *state.State) *Network {
	return &Network{
		stat: stat,
	}
}

func nodePortCreate() (err error) {
	err = iproute.BridgeAdd("", settings.Hypervisor.NodePortNetworkName)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "link", "set",
		"dev", settings.Hypervisor.NodePortNetworkName, "up",
	)
	if err != nil {
		return
	}

	bridges.ClearCache()

	return
}

func nodePortGetAddr() (addr string, err error) {
	address, _, err := iproute.AddressGetIface(
		"", settings.Hypervisor.NodePortNetworkName)
	if err != nil {
		return
	}

	if address != nil {
		addr = address.Local + fmt.Sprintf("/%d", address.Prefix)
	}

	return
}

func nodePortSetAddr(addr string) (err error) {
	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "link", "set",
		"dev", settings.Hypervisor.NodePortNetworkName, "up",
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "addr", "flush",
		"dev", settings.Hypervisor.NodePortNetworkName,
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "addr", "add", addr,
		"dev", settings.Hypervisor.NodePortNetworkName,
	)
	if err != nil {
		return
	}

	return
}

func nodePortClearAddr() (err error) {
	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "link", "set",
		"dev", settings.Hypervisor.NodePortNetworkName, "up",
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "addr", "flush",
		"dev", settings.Hypervisor.NodePortNetworkName,
	)
	if err != nil {
		return
	}

	return
}

func nodePortRemoveNetwork(stat *state.State) (err error) {
	if nodePortCurGateway != "" || stat.HasInterfaces(
		settings.Hypervisor.HostNetworkName) {

		err = nodePortClearAddr()
		if err != nil {
			return
		}

		nodePortCurGateway = ""
	}

	return
}

func NodePortApplyState(stat *state.State) (err error) {
	if !nodePortInitialized {
		addr, e := nodePortGetAddr()
		if e != nil {
			err = e
			return
		}

		nodePortInitialized = true
		nodePortCurGateway = addr
	}

	if !stat.HasInterfaces(settings.Hypervisor.NodePortNetworkName) {
		logrus.WithFields(logrus.Fields{
			"iface": settings.Hypervisor.NodePortNetworkName,
		}).Info("nodeport: Creating node port interface")

		err = nodePortCreate()
		if err != nil {
			return
		}
	}

	nodePortBlock, err := block.GetNodePortBlock(stat.Node().Id)
	if err != nil {
		return
	}

	gatewayCidr := nodePortBlock.GetGatewayCidr()
	if gatewayCidr == "" {
		logrus.WithFields(logrus.Fields{
			"node_port_block": nodePortBlock.Id.Hex(),
		}).Error("nodeport: Node port network block gateway is invalid")

		err = nodePortRemoveNetwork(stat)
		if err != nil {
			return
		}

		return
	}

	if nodePortCurGateway != gatewayCidr {
		logrus.WithFields(logrus.Fields{
			"node_port_block":         nodePortBlock.Id.Hex(),
			"node_port_block_gateway": gatewayCidr,
		}).Info("nodeport: Updating node port network bridge")

		err = nodePortSetAddr(gatewayCidr)
		if err != nil {
			return
		}

		nodePortCurGateway = gatewayCidr
	}

	return
}

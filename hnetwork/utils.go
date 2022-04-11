package hnetwork

import (
	"fmt"

	"github.com/pritunl/pritunl-cloud/iproute"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/utils"
)

func create() (err error) {
	err = iproute.BridgeAdd("", settings.Hypervisor.HostNetworkName)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "link", "set",
		"dev", settings.Hypervisor.HostNetworkName, "up",
	)
	if err != nil {
		return
	}

	return
}

func getAddr() (addr string, err error) {
	address, _, err := iproute.AddressGetIface(
		"", settings.Hypervisor.HostNetworkName)
	if err != nil {
		return
	}

	if address != nil {
		addr = address.Local + fmt.Sprintf("/%d", address.Prefix)
	}

	return
}

func setAddr(addr string) (err error) {
	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "link", "set",
		"dev", settings.Hypervisor.HostNetworkName, "up",
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "addr", "flush",
		"dev", settings.Hypervisor.HostNetworkName,
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "addr", "add", addr,
		"dev", settings.Hypervisor.HostNetworkName,
	)
	if err != nil {
		return
	}

	return
}

func clearAddr() (err error) {
	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "link", "set",
		"dev", settings.Hypervisor.HostNetworkName, "up",
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "addr", "flush",
		"dev", settings.Hypervisor.HostNetworkName,
	)
	if err != nil {
		return
	}

	return
}

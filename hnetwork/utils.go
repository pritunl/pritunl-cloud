package hnetwork

import (
	"github.com/pritunl/pritunl-cloud/iproute"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/utils"
)

func getAddr() (addr string, err error) {
	address, _, err := iproute.AddressGetIface(
		"", settings.Hypervisor.HostNetworkName)
	if err != nil {
		return
	}

	if address != nil {
		addr = address.Local
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

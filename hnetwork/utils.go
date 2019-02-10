package hnetwork

import (
	"net"
	"strings"

	"github.com/pritunl/pritunl-cloud/settings"

	"github.com/pritunl/pritunl-cloud/utils"
)

func getAddr() (addr string, err error) {
	ipData, err := utils.ExecCombinedOutputLogged(
		[]string{
			"No such file or directory",
			"does not exist",
		},
		"ip", "-f", "inet", "-o", "addr",
		"show", "dev", settings.Hypervisor.HostNetworkName,
	)
	if err != nil {
		return
	}

	for _, line := range strings.Split(ipData, "\n") {
		if !strings.Contains(line, "global") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) > 3 {
			ipAddr := net.ParseIP(strings.Split(fields[3], "/")[0])
			if ipAddr != nil && len(ipAddr) > 0 {
				addr = ipAddr.String()
			}
		}

		break
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

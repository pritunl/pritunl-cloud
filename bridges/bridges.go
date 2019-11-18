package bridges

import (
	"net"
	"strings"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/iproute"
	"github.com/pritunl/pritunl-cloud/utils"
)

var (
	bridges         = []string{}
	lastBridgesSync time.Time
	curAddr         = map[string]string{}
	curAddr6        = map[string]string{}
	lastAddrSync    = map[string]time.Time{}
)

func GetBridges() (brdgs []string, err error) {
	if time.Since(lastBridgesSync) < 300*time.Second {
		brdgs = bridges
		return
	}

	bridgesNew := []string{}

	ifaces, err := iproute.IfaceGetBridges("")
	if err != nil {
		return
	}

	for _, iface := range ifaces {
		if iface.Name == "pritunlhost0" {
			continue
		}

		bridgesNew = append(bridgesNew, iface.Name)
	}

	bridges = bridgesNew
	lastBridgesSync = time.Now()
	brdgs = bridgesNew

	return
}

func GetIpAddrs(iface string) (addr string, addr6 string, err error) {
	if time.Since(lastAddrSync[iface]) < 600*time.Second {
		addr = curAddr[iface]
		addr6 = curAddr6[iface]
		return
	}

	if iface == "" {
		err = &errortypes.NotFoundError{
			errors.New("bridges: Invalid external node interface"),
		}
		return
	}

	ipData, err := utils.ExecCombinedOutputLogged(
		[]string{
			"No such file or directory",
			"does not exist",
		},
		"ip", "-f", "inet", "-o", "addr",
		"show", "dev", iface,
	)
	if err != nil {
		return
	}

	if strings.Contains(ipData, "not exist") {
		err = &errortypes.NotFoundError{
			errors.New("bridges: Failed to find external node interface"),
		}
		return
	}

	fields := strings.Fields(ipData)
	if len(fields) > 3 {
		ipAddr := net.ParseIP(strings.Split(fields[3], "/")[0])
		if ipAddr != nil && len(ipAddr) > 0 {
			addr = ipAddr.String()
		}
	}

	ipData, err = utils.ExecCombinedOutputLogged(
		[]string{
			"No such file or directory",
			"does not exist",
		},
		"ip", "-f", "inet6", "-o", "addr",
		"show", "dev", iface,
	)
	if err != nil {
		return
	}

	if !strings.Contains(ipData, "not exist") {
		for _, line := range strings.Split(ipData, "\n") {
			if !strings.Contains(line, "global") {
				continue
			}

			fields = strings.Fields(ipData)
			if len(fields) > 3 {
				ipAddr := net.ParseIP(strings.Split(fields[3], "/")[0])
				if ipAddr != nil && len(ipAddr) > 0 {
					addr6 = ipAddr.String()
				}
			}

			break
		}
	}

	curAddr[iface] = addr
	curAddr6[iface] = addr6
	lastAddrSync[iface] = time.Now()

	return
}

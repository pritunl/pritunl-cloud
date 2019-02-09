package bridges

import (
	"net"
	"strings"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
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

	output, err := utils.ExecOutput("", "brctl", "show")
	if err != nil {
		return
	}

	for i, line := range strings.Split(output, "\n") {
		if i == 0 || strings.HasPrefix(line, " ") ||
			strings.HasPrefix(line, "	") {

			continue
		}

		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) == 0 {
			continue
		}

		if len(fields[0]) == 14 {
			continue
		}

		bridgesNew = append(bridgesNew, fields[0])
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

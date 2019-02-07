package node

import (
	"strings"
	"time"

	"github.com/pritunl/pritunl-cloud/utils"
)

var (
	netIfaces         = []string{}
	netLastIfacesSync time.Time
)

func GetInterfaces() (ifaces []string, err error) {
	if time.Since(netLastIfacesSync) < 15*time.Second {
		ifaces = netIfaces
		return
	}

	ifacesNew := []string{}
	allIfaces, err := utils.GetInterfaces()
	if err != nil {
		return
	}

	for _, iface := range allIfaces {
		if len(iface) == 14 || iface == "lo" ||
			strings.Contains(iface, "br") {

			continue
		}
		ifacesNew = append(ifacesNew, iface)
	}

	ifaces = ifacesNew
	netLastIfacesSync = time.Now()
	netIfaces = ifacesNew

	return
}

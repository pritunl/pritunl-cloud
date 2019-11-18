package bridges

import (
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/iproute"
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

	address, address6, err := iproute.AddressGetIface("", iface)
	if err != nil {
		return
	}

	if address != nil {
		addr = address.Local
	}

	if address6 != nil {
		addr6 = address6.Local
	}

	curAddr[iface] = addr
	curAddr6[iface] = addr6
	lastAddrSync[iface] = time.Now()

	return
}

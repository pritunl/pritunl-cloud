package bridges

import (
	"io/ioutil"
	"strings"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/ip"
	"github.com/pritunl/pritunl-cloud/iproute"
	"github.com/pritunl/pritunl-cloud/utils"
)

var (
	bridges         = []ip.Interface{}
	lastBridgesSync time.Time
	curAddr         = map[string]string{}
	curAddr6        = map[string]string{}
	lastAddrSync    = map[string]time.Time{}
)

func ClearCache() {
	lastBridgesSync = time.Time{}
	bridges = []ip.Interface{}
}

func GetBridges() (brdgs []ip.Interface, err error) {
	if time.Since(lastBridgesSync) < 300*time.Second {
		brdgs = bridges
		return
	}

	bridgesNew := []ip.Interface{}
	bridgesSet := set.NewSet()

	ifaces, err := iproute.IfaceGetBridges("")
	if err != nil {
		return
	}

	ifacesData, err := ip.GetIfacesCached("")
	if err != nil {
		return
	}

	for _, iface := range ifaces {
		if iface.Name == "" {
			continue
		}

		ifaceData := ifacesData[iface.Name]
		if ifaceData != nil {
			bridgesNew = append(bridgesNew, ip.Interface{
				Name:    iface.Name,
				Address: ifaceData.GetAddress(),
			})
		} else {
			bridgesNew = append(bridgesNew, ip.Interface{
				Name: iface.Name,
			})
		}
		bridgesSet.Add(iface.Name)
	}

	exists, err := utils.ExistsDir("/etc/sysconfig/network-scripts")
	if err != nil {
		return
	}

	if exists {
		items, e := ioutil.ReadDir("/etc/sysconfig/network-scripts")
		if e != nil {
			err = &errortypes.ReadError{
				errors.Wrap(e, "bridges: Failed to read network scripts"),
			}
			return
		}

		for _, item := range items {
			name := item.Name()

			if !strings.HasPrefix(name, "ifcfg-") ||
				!strings.Contains(name, ":") {

				continue
			}

			name = name[6:]
			names := strings.Split(name, ":")
			if len(names) != 2 || names[0] == "" {
				continue
			}

			if bridgesSet.Contains(names[0]) && !bridgesSet.Contains(name) {
				bridgesNew = append(bridgesNew, ip.Interface{
					Name: name,
				})
				bridgesSet.Add(name)
			}
		}
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

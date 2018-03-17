package bridge

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/utils"
	"io/ioutil"
	"net"
	"strings"
	"time"
)

const ifaceConf = `TYPE="Ethernet"
BOOTPROTO="none"
NAME="%s"
DEVICE="%s"
ONBOOT="yes"
BRIDGE="%s"
`

var (
	configured = false
	BridgeName = ""
)

func Configure() (err error) {
	if configured {
		return
	}

	bridgeName := settings.Hypervisor.BridgeName

	ifaces, err := net.Interfaces()
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "bridge: Failed to get system interfaces"),
		}
		return
	}

	for _, iface := range ifaces {
		if iface.Name == settings.Hypervisor.BridgeName ||
			iface.Name == "br0" {

			BridgeName = iface.Name
			configured = true
			break
		}
	}

	if configured {
		_, err = utils.ExecCombinedOutputLogged(
			nil, "sysctl", "-w",
			fmt.Sprintf("net.ipv6.conf.%s.accept_ra=2", BridgeName),
		)
		if err != nil {
			return
		}

		return
	}

	defaultIface, err := getDefault()
	if err != nil {
		return
	}

	logrus.WithFields(logrus.Fields{
		"bridge":  bridgeName,
		"gateway": defaultIface,
	}).Info("bridge: Creating bridge interface")

	defaultIfacePath := fmt.Sprintf(
		"/etc/sysconfig/network-scripts/ifcfg-%s",
		defaultIface,
	)
	defaultIfacePathOrig := defaultIfacePath + ".orig"
	bridgeIfacePath := fmt.Sprintf(
		"/etc/sysconfig/network-scripts/ifcfg-%s",
		bridgeName,
	)

	exists, err := utils.Exists(defaultIfacePathOrig)
	if err != nil {
		return
	}

	if !exists {
		err = utils.Exec("",
			"cp", "-f",
			defaultIfacePath, defaultIfacePathOrig,
		)
		if err != nil {
			return
		}
	}

	defaultIfaceConfByt, err := ioutil.ReadFile(defaultIfacePathOrig)
	if err != nil {
		err = &errortypes.ReadError{
			errors.New("bridge: Failed to read network interface"),
		}
		return
	}
	defaultIfaceConf := string(defaultIfaceConfByt)

	bridgeIfaceConf := strings.Replace(
		defaultIfaceConf, "Ethernet", "Bridge", 1)
	bridgeIfaceConf = strings.Replace(
		bridgeIfaceConf, defaultIface, bridgeName, -1)

	defaultIfaceConf = fmt.Sprintf(
		ifaceConf, defaultIface, defaultIface, bridgeName)

	err = ioutil.WriteFile(defaultIfacePath, []byte(defaultIfaceConf), 0644)
	if err != nil {
		err = &errortypes.WriteError{
			errors.New("bridge: Failed to write network interface"),
		}
		return
	}

	err = ioutil.WriteFile(bridgeIfacePath, []byte(bridgeIfaceConf), 0644)
	if err != nil {
		err = &errortypes.WriteError{
			errors.New("bridge: Failed to write network interface"),
		}
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil, "systemctl", "restart", "network")
	if err != nil {
		return
	}

	time.Sleep(5 * time.Second)

	//err = utils.Exec("", "brctl", "addbr", bridgeName)
	//if err != nil {
	//	return
	//}
	//
	//_, err = utils.ExecCombinedOutputLogged(
	//	nil, "sysctl", "-w",
	//	fmt.Sprintf("net.ipv6.conf.%s.accept_ra=2", bridgeName),
	//)
	//if err != nil {
	//	return
	//}
	//
	//err = utils.Exec("", "brctl", "addif", bridgeName, defaultIface)
	//if err != nil {
	//	utils.Exec("", "ip", "link", "set", "dev", bridgeName, "down")
	//	utils.Exec("", "brctl", "delbr", bridgeName)
	//	return
	//}
	//
	//err = utils.Exec("", "ip", "link", "set", "dev",
	//	bridgeName, "up")
	//if err != nil {
	//	utils.Exec("", "brctl", "delif", bridgeName, defaultIface)
	//	utils.Exec("", "ip", "link", "set", "dev", bridgeName, "down")
	//	utils.Exec("", "brctl", "delbr", bridgeName)
	//	return
	//}
	//
	//err = utils.Exec("", "dhclient", bridgeName)
	//if err != nil {
	//	err = nil
	//	utils.Exec("", "dhcpcd", bridgeName)
	//}
	//
	//for i := 0; i < 15; i++ {
	//	ifaces, e := net.Interfaces()
	//	if e != nil {
	//		err = &errortypes.ReadError{
	//			errors.Wrap(e, "bridge: Failed to get system interfaces"),
	//		}
	//		return
	//	}
	//
	//	for _, iface := range ifaces {
	//		if iface.Name == bridgeName {
	//			addrs, e := iface.Addrs()
	//			if e != nil {
	//				err = &errortypes.ReadError{
	//					errors.Wrap(e,
	//						"bridge: Failed to get bridge addresses"),
	//				}
	//				return
	//			}
	//
	//			for _, addr := range addrs {
	//				if !strings.Contains(addr.String(), ":") {
	//					BridgeName = iface.Name
	//					configured = true
	//					return
	//				}
	//			}
	//
	//			break
	//		}
	//	}
	//
	//	time.Sleep(1 * time.Second)
	//}
	//
	//utils.Exec("", "brctl", "delif", bridgeName, defaultIface)
	//utils.Exec("", "ip", "link", "set", "dev", bridgeName, "down")
	//utils.Exec("", "brctl", "delbr", bridgeName)
	//
	//err = &errortypes.ReadError{
	//	errors.Wrap(err, "bridge: Bridge dhcp timeout"),
	//}

	configured = true

	return
}

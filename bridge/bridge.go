package bridge

import (
	"github.com/Sirupsen/logrus"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/utils"
	"net"
	"strings"
	"time"
)

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
			return
		}
	}

	defaultIface, err := getDefault()
	if err != nil {
		return
	}

	logrus.WithFields(logrus.Fields{
		"bridge":  bridgeName,
		"gateway": defaultIface,
	}).Info("bridge: Creating bridge interface")

	err = utils.Exec("", "brctl", "addbr", bridgeName)
	if err != nil {
		return
	}

	err = utils.Exec("", "brctl", "addif", bridgeName, defaultIface)
	if err != nil {
		utils.Exec("", "ip", "link", "set", "dev", bridgeName, "down")
		utils.Exec("", "brctl", "delbr", bridgeName)
		return
	}

	err = utils.Exec("", "ip", "link", "set", "dev",
		bridgeName, "up")
	if err != nil {
		utils.Exec("", "brctl", "delif", bridgeName, defaultIface)
		utils.Exec("", "ip", "link", "set", "dev", bridgeName, "down")
		utils.Exec("", "brctl", "delbr", bridgeName)
		return
	}

	err = utils.Exec("", "dhclient", bridgeName)
	if err != nil {
		err = nil
		utils.Exec("", "dhcpcd", bridgeName)
	}

	for i := 0; i < 15; i++ {
		ifaces, e := net.Interfaces()
		if e != nil {
			err = &errortypes.ReadError{
				errors.Wrap(e, "bridge: Failed to get system interfaces"),
			}
			return
		}

		for _, iface := range ifaces {
			if iface.Name == bridgeName {
				addrs, e := iface.Addrs()
				if e != nil {
					err = &errortypes.ReadError{
						errors.Wrap(e,
							"bridge: Failed to get bridge addresses"),
					}
					return
				}

				for _, addr := range addrs {
					if !strings.Contains(addr.String(), ":") {
						BridgeName = iface.Name
						configured = true
						return
					}
				}

				break
			}
		}

		time.Sleep(1 * time.Second)
	}

	utils.Exec("", "brctl", "delif", bridgeName, defaultIface)
	utils.Exec("", "ip", "link", "set", "dev", bridgeName, "down")
	utils.Exec("", "brctl", "delbr", bridgeName)

	err = &errortypes.ReadError{
		errors.Wrap(err, "bridge: Bridge dhcp timeout"),
	}

	return
}

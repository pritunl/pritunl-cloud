package bridge

import (
	"net"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/settings"
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
)

func Configured() bool {
	return configured
}

func configureBridge() (err error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "bridge: Failed to get system interfaces"),
		}
		return
	}

	for _, iface := range ifaces {
		if iface.Name == settings.Hypervisor.BridgeName ||
			strings.Contains(iface.Name, "pritunlbr") ||
			iface.Name == "br0" {

			settings.Local.BridgeName = iface.Name
			return
		}
	}

	defaultIface, err := getDefault()
	if err != nil {
		return
	}

	if strings.Contains(defaultIface, "br") {
		settings.Local.BridgeName = defaultIface
		return
	}

	logrus.WithFields(logrus.Fields{
		"gateway": defaultIface,
	}).Warn("bridge: No bridge interface found")

	return
}

func Configure() (err error) {
	if configured {
		return
	}

	err = configureBridge()
	if err != nil {
		return
	}

	configured = true
	return
}

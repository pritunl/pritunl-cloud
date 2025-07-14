package cmd

import (
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/dhcpc"
	"github.com/pritunl/pritunl-cloud/dhcps"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/sirupsen/logrus"
)

func Dhcp4Server() (err error) {
	config := strings.Trim(os.Getenv("CONFIG"), "'")
	server4 := &dhcps.Server4{}

	err = json.Unmarshal([]byte(config), server4)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "cmd: Failed to parse DHCP4 configuration"),
		}
		return
	}

	for {
		err = server4.Start()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("cmd: DHCP4 server error")
		}

		time.Sleep(3 * time.Second)
	}
}

func Dhcp4Client() (err error) {
	err = dhcpc.Main()
	if err != nil {
		return
	}

	return
}

func Dhcp6Server() (err error) {
	config := strings.Trim(os.Getenv("CONFIG"), "'")
	server6 := &dhcps.Server6{}

	err = json.Unmarshal([]byte(config), server6)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "cmd: Failed to parse DHCP6 configuration"),
		}
		return
	}

	for {
		err = server6.Start()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("cmd: DHCP6 server error")
		}

		time.Sleep(3 * time.Second)
	}
}

func NdpServer() (err error) {
	config := strings.Trim(os.Getenv("CONFIG"), "'")
	serverNdp := &dhcps.ServerNdp{}

	err = json.Unmarshal([]byte(config), serverNdp)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "cmd: Failed to parse NDP configuration"),
		}
		return
	}

	for {
		err = serverNdp.Start()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("cmd: NDP server error")
		}

		time.Sleep(3 * time.Second)
	}
}

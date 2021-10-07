package systemctl

import (
	"github.com/pritunl/pritunl-cloud/builder/prompt"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/sirupsen/logrus"
)

func Firewall() (err error) {
	resp, err := prompt.ConfirmDefault(
		"Disable firewalld [Y/n]",
		true,
	)
	if err != nil {
		return
	}

	if !resp {
		return
	}

	err = utils.Exec("", "/usr/bin/yum", "-y", "remove", "iptables-services")
	if err != nil {
		return
	}

	utils.ExecCombinedOutput("/usr/bin/systemctl", "disable", "firewalld")
	utils.ExecCombinedOutput("/usr/bin/systemctl", "stop", "firewalld")

	logrus.WithFields(logrus.Fields{
		"service": "firewalld",
	}).Info("systemctl: Firewalld disabled")

	return
}

func Systemctl() (err error) {
	err = Firewall()
	if err != nil {
		return
	}

	return
}

package start

import (
	"github.com/pritunl/pritunl-cloud/builder/prompt"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/sirupsen/logrus"
)

func Systemctl() (err error) {
	_, err = utils.ExecOutputLogged(nil,
		"/usr/bin/systemctl", "start", "pritunl-cloud")
	if err != nil {
		return
	}

	logrus.WithFields(logrus.Fields{
		"service": "pritunl-cloud",
	}).Info("start: Pritunl Cloud started")

	return
}

func Start(noStart bool) (err error) {
	if noStart {
		return
	}

	resp, err := prompt.ConfirmDefault(
		"Start Pritunl Cloud [Y/n]",
		true,
	)
	if err != nil {
		return
	}

	if !resp {
		return
	}

	err = Systemctl()
	if err != nil {
		return
	}

	return
}

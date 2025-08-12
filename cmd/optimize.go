package cmd

import (
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/sirupsen/logrus"
)

func Optimize() (err error) {
	err = optimizeNested()
	if err != nil {
		return
	}

	return
}

func optimizeNested() (err error) {
	resp, err := utils.ConfirmDefault(
		"Enable nested virtualization",
		true,
	)
	if err != nil {
		return
	}

	pth := "/etc/modprobe.d/kvm-nested.conf"
	if resp {
		err = utils.CreateWrite(
			pth,
			"options kvm-intel nested=1\noptions kvm-amd nested=1\n",
			0644,
		)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"path": pth,
		}).Info("sysctl: Nested virtualization enabled")
	} else {
		exists, e := utils.Exists(pth)
		if e != nil {
			err = e
			return
		}

		if exists {
			err = utils.Remove(pth)
			if err != nil {
				return
			}

			logrus.WithFields(logrus.Fields{
				"path": pth,
			}).Info("sysctl: Nested virtualization disabled")
		}
	}

	return
}

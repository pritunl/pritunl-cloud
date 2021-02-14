package sysctl

import (
	"github.com/pritunl/pritunl-cloud/builder/prompt"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/sirupsen/logrus"
)

const (
	intelNested = `options kvm-intel nested=1
`
	amdNested = `options kvm-amd nested=1
`
)

func Nested() (err error) {
	resp, err := prompt.ConfirmDefault(
		"Enable nested virtualization [Y/n]",
		true,
	)
	if err != nil {
		return
	}

	pth := "/etc/modprobe.d/kvm-intel.conf"
	if resp {
		err = utils.CreateWrite(
			pth,
			intelNested,
			0644,
		)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"path": pth,
		}).Info("sysctl: Nested Intel virtualization enabled")
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
			}).Info("sysctl: Nested Intel virtualization disabled")
		}
	}

	pth = "/etc/modprobe.d/kvm-amd.conf"
	if resp {
		err = utils.CreateWrite(
			pth,
			amdNested,
			0644,
		)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"path": pth,
		}).Info("sysctl: Nested AMD virtualization enabled")
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
			}).Info("sysctl: Nested AMD virtualization disabled")
		}
	}

	return
}

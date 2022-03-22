package cmd

import (
	"github.com/pritunl/pritunl-cloud/mtu"
)

func MtuCheck() (err error) {
	chk := mtu.NewCheck()

	err = chk.Run()
	if err != nil {
		return
	}

	return
}

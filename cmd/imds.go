package cmd

import (
	"github.com/pritunl/pritunl-cloud/imds/server"
)

func ImdsServer() (err error) {
	err = server.Main()
	if err != nil {
		return
	}

	return
}

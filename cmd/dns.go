package cmd

import (
	"github.com/pritunl/pritunl-cloud/dnss"
)

func DnsServer() (err error) {
	dnss.Run()
	return
}

package sync

import (
	"github.com/pritunl/pritunl-cloud/ipsec"
)

func initIpsec() {
	go ipsec.RunSync()
}

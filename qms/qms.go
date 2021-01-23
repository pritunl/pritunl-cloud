package qms

import (
	"time"

	"github.com/pritunl/pritunl-cloud/utils"
)

var (
	socketsLock = utils.NewMultiTimeoutLock(1 * time.Minute)
)

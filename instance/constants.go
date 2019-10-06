package instance

import (
	"github.com/dropbox/godropbox/container/set"
)

const (
	Provision = "provision"
	Start     = "start"
	Stop      = "stop"
	Cleanup   = "cleanup"
	Restart   = "restart"
	Destroy   = "destroy"
)

var (
	ValidStates = set.NewSet(
		Provision,
		Start,
		Stop,
		Cleanup,
		Restart,
		Destroy,
	)
)

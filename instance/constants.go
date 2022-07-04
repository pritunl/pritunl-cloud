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
	Linux     = "linux"
	BSD       = "bsd"
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
	ValidCloudTypes = set.NewSet(
		Linux,
		BSD,
	)
)

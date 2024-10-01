package plan

import (
	"github.com/dropbox/godropbox/container/set"
)

const (
	Start   = "start"
	Stop    = "stop"
	Restart = "restart"
	Destroy = "destroy"
)

var actions = set.NewSet(
	Start,
	Stop,
	Restart,
	Destroy,
)

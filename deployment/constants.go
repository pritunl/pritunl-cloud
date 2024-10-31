package deployment

import (
	"time"
)

const (
	Reserved = "reserved"
	Deployed = "deployed"
	Destroy  = "destroy"

	Instance = "instance"

	ThresholdMin = 10
	ActionLimit  = 1 * time.Minute
)

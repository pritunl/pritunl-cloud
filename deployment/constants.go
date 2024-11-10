package deployment

import (
	"time"
)

const (
	Reserved = "reserved"
	Deployed = "deployed"
	Destroy  = "destroy"
	Archive  = "archive"
	Archived = "archived"
	Restore  = "restore"

	Instance = "instance"

	Healthy   = "healthy"
	Unhealthy = "unhealthy"

	ThresholdMin = 10
	ActionLimit  = 1 * time.Minute
)

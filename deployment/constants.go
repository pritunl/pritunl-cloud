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

	Ready    = "ready"
	Snapshot = "snapshot"
	Complete = "complete"
	Failed   = "failed"

	Instance = "instance"
	Image    = "image"
	Firewall = "firewall"
	Domain   = "domain"

	Healthy   = "healthy"
	Unhealthy = "unhealthy"

	ThresholdMin = 10
	ActionLimit  = 1 * time.Minute
)

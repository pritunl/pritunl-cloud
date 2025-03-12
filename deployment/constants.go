package deployment

import (
	"time"
)

const (
	Provision = "provision"
	Reserved  = "reserved"
	Deployed  = "deployed"
	Archived  = "archived"

	Destroy = "destroy"
	Archive = "archive"
	Migrate = "migrate"
	Restore = "restore"

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

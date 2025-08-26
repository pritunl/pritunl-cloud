package dhcpc

import (
	"time"
)

const (
	MaxMessageSize  = 1500
	DefaultInterval = 60 * time.Second
	DhcpTimeout     = 10 * time.Second
	DhcpRetries     = 3
)

package dhcpc

import (
	"time"
)

const (
	MaxMessageSize = 1500
	DhcpTimeout    = 10 * time.Second
	DhcpRetries    = 3
	PreferredTtl   = 24 * time.Hour
)

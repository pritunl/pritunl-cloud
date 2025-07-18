package dhcpc

import (
	"time"
)

const (
	MaxMessageSize = 1500
	DhcpTimeout    = 10 * time.Second
	DhcpRetries    = 3
)

var (
	ImdsAddress = ""
	ImdsPort    = ""
	ImdsSecret  = ""
	DhcpIface   = ""
	DhcpIface6  = ""
)

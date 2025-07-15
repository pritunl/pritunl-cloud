package dhcpc

import (
	"time"
)

const (
	MaxMessageSize = 1500
	DhcpTimeout    = 30 * time.Second
	DhcpRetries    = 3
)

var (
	ImdsAddress = ""
	ImdsPort    = ""
	DhcpSecret  = ""
	DhcpIface   = ""
	DhcpIface6  = ""
)

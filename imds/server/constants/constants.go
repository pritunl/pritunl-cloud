package constants

import (
	"time"
)

const (
	Version     = "1.0.3229.20"
	ConfRefresh = 500 * time.Millisecond
)

var (
	Sock         = ""
	Host         = "127.0.0.1"
	Port         = 80
	Client       = "127.0.0.1"
	ClientSecret = ""
	DhcpSecret   = ""
	HostSecret   = ""
	Interrupt    = false
)

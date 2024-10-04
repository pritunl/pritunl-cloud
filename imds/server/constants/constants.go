package constants

import (
	"time"
)

const (
	Version       = "1.0.3229.20"
	Authenticated = false
	AuthKey       = "test"
	ConfRefresh   = 500 * time.Millisecond
)

var (
	Host      = "127.0.0.1"
	Port      = 80
	Interrupt = false
)

package constants

import (
	"time"
)

const (
	Version         = "1.0.800.79"
	DatabaseVersion = 1
	ConfPath        = "/etc/pritunl-cloud.json"
	LogPath         = "/var/log/pritunl-cloud.log"
	LogPath2        = "/var/log/pritunl-cloud.log.1"
	TempPath        = "/tmp/pritunl-cloud"
	StaticCache     = true
	RetryDelay      = 3 * time.Second
)

var (
	Production = true
	Interrupt  = false
	StaticRoot = []string{
		"www/dist",
		"/usr/share/pritunl-cloud/www",
	}
	StaticTestingRoot = []string{
		"www",
		"/usr/share/pritunl-cloud/www",
	}
)

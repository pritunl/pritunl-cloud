package constants

import (
	"go/build"
	"path"
	"time"
)

const (
	Version         = "1.0.800.79"
	DatabaseVersion = 1
	ConfPath        = "/cloud/pritunl-cloud.json"
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
		path.Join(
			build.Default.GOPATH,
			"src/github.com/pritunl/pritunl-cloud/www/dist",
		),
		"/home/cloud/go/src/github.com/pritunl/pritunl-cloud/www/dist",
	}
	StaticTestingRoot = []string{
		"www",
		"/usr/share/pritunl-cloud/www",
		path.Join(
			build.Default.GOPATH,
			"src/github.com/pritunl/pritunl-cloud/www",
		),
		"/home/cloud/go/src/github.com/pritunl/pritunl-cloud/www",
	}
)

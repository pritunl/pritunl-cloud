package constants

import (
	"time"
)

const (
	Version         = "2.0.3615.44"
	DatabaseVersion = 1
	LogPath         = "/var/log/pritunl-cloud.log"
	LogPath2        = "/var/log/pritunl-cloud.log.1"
	StaticCache     = true
	RetryDelay      = 3 * time.Second
)

var (
	Production   = true
	DebugWeb     = false
	FastExit     = false
	LockDebug    = false
	Interrupt    = false
	Shutdown     = false
	ConfPath     = "/etc/pritunl-cloud.json"
	DefaultRoot  = "/var/lib/pritunl-cloud"
	DefaultCache = "/var/cache/pritunl-cloud"
	DefaultTemp  = "/tmp/pritunl-cloud"
	StaticRoot   = []string{
		"www/dist",
		"/usr/share/pritunl-cloud/www",
	}
	StaticTestingRoot = []string{
		"/home/cloud/git/pritunl-cloud/www/dist-dev",
		"/home/cloud/go/src/github.com/pritunl/pritunl-cloud/www/dist-dev",
		"/usr/share/pritunl-cloud/www",
	}
)

var PritunlKeyring = `-----BEGIN PGP PUBLIC KEY BLOCK-----

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
3FxR7SpBSsF/SmguI/w=
=ohOM
-----END PGP PUBLIC KEY BLOCK-----`

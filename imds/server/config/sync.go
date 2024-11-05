package config

import (
	"os"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/imds/server/constants"
	"github.com/pritunl/pritunl-cloud/imds/server/errortypes"
	"github.com/pritunl/tools/logger"
)

var (
	curMod time.Time
)

func GetModTime() (mod time.Time, err error) {
	stat, err := os.Stat(Path)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "config: Failed to stat conf file"),
		}
		return
	}

	mod = stat.ModTime()

	return
}

func SyncConfig() (err error) {
	if constants.Interrupt {
		return
	}

	mod, err := GetModTime()
	if err != nil {
		return
	}

	if mod != curMod {
		time.Sleep(100 * time.Millisecond)

		mod, err = GetModTime()
		if err != nil {
			return
		}

		err = Load()
		if err != nil {
			return
		}

		logger.Info("Reloaded config")

		curMod = mod
	}

	return
}

func runSyncConfig() {
	curMod, _ = GetModTime()

	for {
		time.Sleep(constants.ConfRefresh)

		err := SyncConfig()
		if err != nil {
			logger.WithFields(logger.Fields{
				"error": err,
			}).Info("sync: Failed to sync config")
		}
	}
}

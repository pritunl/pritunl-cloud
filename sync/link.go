package sync

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/pritunl/pritunl-cloud/ipsec"
)

func linkRunner() {
	for {
		err := ipsec.InitState()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("sync: Failed to init ipsec state")

			time.Sleep(2 * time.Second)
			continue
		}

		break
	}

	for {
		time.Sleep(2 * time.Second)

		err := ipsec.SyncState()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("sync: Failed to sync ipsec state")
		}
	}
}

func initLink() {
	go linkRunner()
}

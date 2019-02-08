package sync

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/ipsec"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/vpc"
)

func linkSync() (err error) {
	db := database.GetDatabase()
	defer db.Close()

	vpcs, err := vpc.GetAll(db, &bson.M{})
	if err != nil {
		return
	}

	ipsec.SyncStates(vpcs)

	return
}

func linkRunner() {
	for {
		time.Sleep(2 * time.Second)

		if !node.Self.IsHypervisor() {
			continue
		}

		err := linkSync()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("sync: Failed to sync IPsec links")
			return
		}
	}

	return
}

func initLink() {
	go linkRunner()
}

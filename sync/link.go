package sync

import (
	"github.com/Sirupsen/logrus"
	"github.com/pritunl/pritunl-cloud/bridge"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/ipsec"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/vpc"
	"gopkg.in/mgo.v2/bson"
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
		if !node.Self.IsHypervisor() || !bridge.Configured() {
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

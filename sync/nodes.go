package sync

import (
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/interfaces"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/settings"
)

func nodeSync() (err error) {
	db := database.GetDatabase()
	defer db.Close()

	appId := ""
	facets := []string{}

	if node.Self.UserDomain != "" {
		appId = fmt.Sprintf("https://%s/auth/u2f/app.json",
			node.Self.UserDomain)
	}

	nodes, err := node.GetAll(db)
	if err != nil {
		return
	}

	for _, nde := range nodes {
		if appId == "" {
			appId = fmt.Sprintf("https://%s/auth/u2f/app.json",
				nde.UserDomain)
		}
		if nde.UserDomain != "" {
			facets = append(facets,
				fmt.Sprintf("https://%s", nde.UserDomain))
		}
		if nde.AdminDomain != "" {
			facets = append(facets,
				fmt.Sprintf("https://%s", nde.AdminDomain))
		}
	}

	settings.Local.AppId = appId
	settings.Local.Facets = facets

	interfaces.SyncIfaces(false)

	return
}

func nodeRunner() {
	time.Sleep(1 * time.Second)

	for {
		err := nodeSync()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("sync: Failed to sync node status")
		}

		time.Sleep(3 * time.Second)
	}
}

func initNode() {
	interfaces.SyncIfaces(true)

	go nodeRunner()
}

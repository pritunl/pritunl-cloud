package sync

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/settings"
)

func nodeSync() (err error) {
	db := database.GetDatabase()
	defer db.Close()

	appId := ""
	facets := []string{}

	if node.Self.UserDomain != "" {
		domain := node.Self.UserDomain
		port := node.Self.Port
		if node.Self.Protocol == "https" && port != 443 {
			domain += ":" + strconv.Itoa(port)
		}
		appId = fmt.Sprintf("https://%s/auth/u2f/app.json", domain)
	}

	nodes, err := node.GetAll(db)
	if err != nil {
		return
	}

	domains := set.NewSet()
	for _, nde := range nodes {
		if appId == "" {
			appId = fmt.Sprintf("https://%s/auth/u2f/app.json",
				nde.UserDomain)
		}

		domain := nde.UserDomain
		port := nde.Port
		if domain != "" {
			if nde.Protocol == "https" && port != 443 {
				domain += ":" + strconv.Itoa(port)
			}

			if !domains.Contains(domain) {
				domains.Add(domain)
				facets = append(facets, fmt.Sprintf("https://%s", domain))
			}
		}

		domain = nde.AdminDomain
		port = nde.Port
		if domain != "" {
			if nde.Protocol == "https" && port != 443 {
				domain += ":" + strconv.Itoa(port)
			}

			if !domains.Contains(domain) {
				domains.Add(domain)
				facets = append(facets, fmt.Sprintf("https://%s", domain))
			}
		}
	}

	settings.Local.AppId = appId
	settings.Local.Facets = facets

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
	go nodeRunner()
}

package sync

import (
	"github.com/Sirupsen/logrus"
	"github.com/pritunl/pritunl-cloud/bridge"
	"github.com/pritunl/pritunl-cloud/constants"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/deploy"
	"github.com/pritunl/pritunl-cloud/firewall"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/iptables"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/state"
	"time"
)

func deployState() (err error) {
	stat, err := state.GetState()
	if err != nil {
		return
	}

	err = deploy.Deploy(stat)
	if err != nil {
		return
	}

	return
}

func syncNodeFirewall() {
	db := database.GetDatabase()
	defer db.Close()

	for i := 0; i < 2; i++ {
		fires, err := firewall.GetRoles(db, node.Self.NetworkRoles)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("sync: Failed to get node firewall rules")
			return
		}

		ingress := firewall.MergeIngress(fires)

		err = iptables.UpdateState([]*instance.Instance{}, []string{},
			ingress, map[string][]*firewall.Rule{})
		if err != nil {
			if i < 1 {
				err = nil
				time.Sleep(300 * time.Millisecond)
				continue
			} else {
				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Error("sync: Failed to update iptables, resetting state")
				for {
					err = iptables.Recover()
					if err != nil {
						logrus.WithFields(logrus.Fields{
							"error": err,
						}).Error("sync: Failed to recover iptables, retrying")
						continue
					}
					break
				}
			}
		}

		break
	}
}

func vmRunner() {
	time.Sleep(1 * time.Second)

	for {
		time.Sleep(1 * time.Second)
		if !node.Self.IsHypervisor() {
			syncNodeFirewall()
			continue
		}

		err := bridge.Configure()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("sync: Failed to configure bridge")

			time.Sleep(1 * time.Second)

			continue
		}

		break
	}

	logrus.WithFields(logrus.Fields{
		"production": constants.Production,
		"bridge":     settings.Local.BridgeName,
	}).Info("sync: Starting hypervisor")

	for {
		time.Sleep(2500 * time.Millisecond)
		if !node.Self.IsHypervisor() {
			syncNodeFirewall()
			break
		}

		err := deployState()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("sync: Failed to deploy state")
			continue
		}
	}
}

func initVm() {
	go vmRunner()
}

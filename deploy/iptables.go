package deploy

import (
	"github.com/sirupsen/logrus"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/iptables"
	"github.com/pritunl/pritunl-cloud/state"
)

type Iptables struct {
	stat *state.State
}

func (t *Iptables) Deploy() (err error) {
	db := database.GetDatabase()
	defer db.Close()

	nodeSelf := t.stat.Node()
	instaces := t.stat.Instances()
	namespaces := t.stat.Namespaces()
	nodeFirewall := t.stat.NodeFirewall()
	firewalls := t.stat.Firewalls()

	err = iptables.UpdateState(nodeSelf, instaces, namespaces,
		nodeFirewall, firewalls)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("deploy: Failed to update iptables, resetting state")
		for {
			err = iptables.Recover()
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Error("deploy: Failed to recover iptables, retrying")
				continue
			}
			break
		}
		err = nil
		return
	}

	return
}

func NewIptables(stat *state.State) *Iptables {
	return &Iptables{
		stat: stat,
	}
}

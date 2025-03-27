package iptables

import (
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/firewall"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/ipvs"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vpc"
	"github.com/sirupsen/logrus"
)

type Update struct {
	OldState         *State
	NewState         *State
	Namespaces       []string
	FailedNamespaces set.Set
}

func (u *Update) Apply() {
	changed := false
	var removed []string
	oldIfaces := set.NewSet()
	newIfaces := set.NewSet()

	namespacesSet := set.NewSet()
	for _, namespace := range u.Namespaces {
		namespacesSet.Add(namespace)
	}

	for iface := range u.OldState.Interfaces {
		oldIfaces.Add(iface)
	}
	for iface := range u.NewState.Interfaces {
		newIfaces.Add(iface)
	}

	oldIfaces.Subtract(newIfaces)
	for iface := range oldIfaces.Iter() {
		removed = append(removed, iface.(string))
		err := u.OldState.Interfaces[iface.(string)].Remove(nil)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"iface": iface,
				"error": err,
			}).Error("iptables: Failed to delete removed interface iptables")
		}
	}

	if removed != nil {
		logrus.WithFields(logrus.Fields{
			"ifaces": removed,
		}).Info("iptables: Removed iptables")
	}

	for _, rules := range u.NewState.Interfaces {
		if u.FailedNamespaces.Contains(rules.Namespace) {
			logrus.WithFields(logrus.Fields{
				"namespace": rules.Namespace,
			}).Warn("iptables: Skipping failed namespace")
			continue
		}

		if rules.Namespace != "0" &&
			!namespacesSet.Contains(rules.Namespace) {

			_, err := utils.ExecCombinedOutputLogged(
				[]string{"File exists"},
				"ip", "netns",
				"add", rules.Namespace,
			)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"namespace": rules.Namespace,
					"error":     err,
				}).Error("iptables: Namespace add error")

				u.FailedNamespaces.Add(rules.Namespace)
				continue
			}
		}

		var diff *RulesDiff
		oldRules := u.OldState.Interfaces[rules.Namespace+"-"+rules.Interface]
		if oldRules != nil {
			diff = diffRules(oldRules, rules)
			if diff == nil {
				continue
			}

			if !changed {
				changed = true
				logrus.WithFields(logrus.Fields{
					"ingress":  diff.IngressDiff,
					"ingress6": diff.Ingress6Diff,
					"nats":     diff.NatsDiff,
					"nats6":    diff.Nats6Diff,
					"maps":     diff.MapsDiff,
					"maps6":    diff.Maps6Diff,
					"holds":    diff.HoldsDiff,
					"holds6":   diff.Holds6Diff,
				}).Info("iptables: Updating iptables")
			}

			if (diff.IngressDiff || diff.Ingress6Diff) &&
				rules.Interface != "host" {

				err := rules.Hold()
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"namespace": rules.Namespace,
						"error":     err,
					}).Error("iptables: Namespace hold error")

					u.FailedNamespaces.Add(rules.Namespace)
					continue
				}
			}

			err := oldRules.Remove(diff)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"namespace": rules.Namespace,
					"error":     err,
				}).Error("iptables: Namespace remove error")

				u.FailedNamespaces.Add(rules.Namespace)
				continue
			}
		}

		if !changed {
			changed = true
			logrus.Info("iptables: Updating iptables")
		}

		err := rules.Apply(diff)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"namespace": rules.Namespace,
				"error":     err,
			}).Error("iptables: Namespace apply error")

			u.FailedNamespaces.Add(rules.Namespace)
			continue
		}
	}

	if u.NewState.Interfaces["0-host"].Ipvs != nil {
		err := ipvs.UpdateState(u.NewState.Interfaces["0-host"].Ipvs)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("storage: Failed to update ipvs state")
		}
	}

	return
}

func (u *Update) Recover() {
	if u.FailedNamespaces.Contains("0") {
		err := RecoverNode()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("deploy: Failed to recover node iptables, retrying")
			time.Sleep(10 * time.Second)
		}
	}

	if u.FailedNamespaces.Len() > 0 {
		logrus.Error("deploy: Failed to update iptables, " +
			"reloading state")

		time.Sleep(10 * time.Second)

		err := u.reload()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("deploy: Failed to recover iptables")
		}
	}
}

func (u *Update) reload() (err error) {
	db := database.GetDatabase()
	defer db.Close()

	nodeDatacenter, err := node.Self.GetDatacenter(db)
	if err != nil {
		return
	}

	vpcs := []*vpc.Vpc{}
	if !nodeDatacenter.IsZero() {
		vpcs, err = vpc.GetDatacenter(db, nodeDatacenter)
		if err != nil {
			return
		}
	}

	namespaces, err := utils.GetNamespaces()
	if err != nil {
		return
	}

	instances, err := instance.GetAllVirt(db, &bson.M{
		"node": node.Self.Id,
	}, nil, nil)
	if err != nil {
		return
	}

	specRules, nodePortsMap, err := firewall.GetSpecRulesSlow(
		db, node.Self.Id, instances)
	if err != nil {
		return
	}

	nodeFirewall, firewalls, firewallMaps, err := firewall.GetAllIngress(
		db, node.Self, instances, specRules, nodePortsMap)
	if err != nil {
		return
	}

	err = Init(namespaces, vpcs, instances, nodeFirewall,
		firewalls, firewallMaps)
	if err != nil {
		return
	}

	return
}

func ApplyUpdate(newState *State, namespaces []string, recover bool) {
	lockId := stateLock.Lock()

	update := &Update{
		OldState:         curState,
		NewState:         newState,
		Namespaces:       namespaces,
		FailedNamespaces: set.NewSet(),
	}

	update.Apply()

	curState = newState

	stateLock.Unlock(lockId)

	if recover {
		update.Recover()
	}

	return
}

func UpdateState(nodeSelf *node.Node, vpcs []*vpc.Vpc,
	instances []*instance.Instance, namespaces []string,
	nodeFirewall []*firewall.Rule, firewalls map[string][]*firewall.Rule,
	firewallMaps map[string][]*firewall.Mapping) {

	newState := LoadState(nodeSelf, vpcs, instances, nodeFirewall,
		firewalls, firewallMaps)

	ApplyUpdate(newState, namespaces, false)

	return
}

func UpdateStateRecover(nodeSelf *node.Node, vpcs []*vpc.Vpc,
	instances []*instance.Instance, namespaces []string,
	nodeFirewall []*firewall.Rule, firewalls map[string][]*firewall.Rule,
	firewallMaps map[string][]*firewall.Mapping) {

	newState := LoadState(nodeSelf, vpcs, instances, nodeFirewall,
		firewalls, firewallMaps)

	ApplyUpdate(newState, namespaces, true)

	return
}

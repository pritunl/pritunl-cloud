package iptables

import (
	"strings"
	"time"

	"github.com/dropbox/godropbox/container/set"
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
	SkipNamespaces   set.Set
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
		rules := u.OldState.Interfaces[iface.(string)]
		if u.SkipNamespaces.Contains(rules.Namespace) {
			continue
		}

		removed = append(removed, iface.(string))
		err := rules.Remove(nil)
		if err != nil {
			u.FailedNamespaces.Add(rules.Namespace)
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
		if u.SkipNamespaces.Contains(rules.Namespace) {
			continue
		}

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

			if rules.Interface == "host" {
				if diff.IngressDiff {
					rules.Holds = append(rules.Holds, oldRules.Holds...)
					oldRules.Holds = [][]string{}
				}
				if diff.Ingress6Diff {
					rules.Holds6 = append(rules.Holds6, oldRules.Holds6...)
					oldRules.Holds6 = [][]string{}
				}
			} else if diff.IngressDiff || diff.Ingress6Diff {
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

	hostIface := u.NewState.Interfaces["0-host"]
	if hostIface != nil && hostIface.Ipvs != nil {
		err := ipvs.UpdateState(u.NewState.Interfaces["0-host"].Ipvs)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("storage: Failed to update ipvs state")
		}
	}

	return
}

func (u *Update) reload(namespace string) (err error) {
	state := &State{
		Interfaces: map[string]*Rules{},
	}

	exists := true
	if namespace != "0" {
		exists, err = utils.Exists("/var/run/netns/" + namespace)
		if err != nil {
			return
		}
	}

	if exists {
		instIface := ""
		for _, st := range []*State{u.OldState, u.NewState} {
			for _, rules := range st.Interfaces {
				if rules.Namespace == namespace &&
					strings.HasPrefix(rules.Interface, "p") {

					instIface = rules.Interface
				}
			}
		}

		err = loadIptables(namespace, instIface, state, false)
		if err != nil {
			return
		}

		err = loadIptables(namespace, instIface, state, true)
		if err != nil {
			return
		}
	}

	for key, rules := range u.OldState.Interfaces {
		if rules.Namespace == namespace {
			delete(u.OldState.Interfaces, key)
		}
	}
	for key, rules := range state.Interfaces {
		u.OldState.Interfaces[key] = rules
	}

	return
}

func (u *Update) Refresh() {
	now := time.Now()

	for namespace, failure := range curFailures {
		if now.Before(failure.Retry) {
			u.SkipNamespaces.Add(namespace)
			continue
		}

		err := u.reload(namespace)
		if err != nil {
			u.SkipNamespaces.Add(namespace)
			delay := failure.Delay()

			logrus.WithFields(logrus.Fields{
				"namespace": namespace,
				"failures":  failure.Count,
				"retry":     delay.String(),
				"error":     err,
			}).Error("iptables: Failed to reload namespace iptables")
		}
	}
}

func (u *Update) Commit() {
	state := &State{
		Interfaces: map[string]*Rules{},
	}

	for key, rules := range u.NewState.Interfaces {
		if !u.SkipNamespaces.Contains(rules.Namespace) &&
			!u.FailedNamespaces.Contains(rules.Namespace) {

			state.Interfaces[key] = rules
		}
	}
	for key, rules := range u.OldState.Interfaces {
		if u.SkipNamespaces.Contains(rules.Namespace) ||
			u.FailedNamespaces.Contains(rules.Namespace) {

			state.Interfaces[key] = rules
		}
	}

	curState = state

	for namespace := range curFailures {
		if !u.SkipNamespaces.Contains(namespace) &&
			!u.FailedNamespaces.Contains(namespace) {

			delete(curFailures, namespace)
		}
	}

	for namespaceInf := range u.FailedNamespaces.Iter() {
		namespace := namespaceInf.(string)

		failure := curFailures[namespace]
		if failure == nil {
			failure = &Failure{}
			curFailures[namespace] = failure
		}
		delay := failure.Delay()

		logrus.WithFields(logrus.Fields{
			"namespace": namespace,
			"failures":  failure.Count,
			"retry":     delay.String(),
		}).Error("iptables: Failed to update namespace iptables")
	}
}

func ApplyUpdate(newState *State, namespaces []string, recover bool) {
	lockId := stateLock.Lock()
	defer stateLock.Unlock(lockId)

	update := &Update{
		OldState:         curState,
		NewState:         newState,
		Namespaces:       namespaces,
		FailedNamespaces: set.NewSet(),
		SkipNamespaces:   set.NewSet(),
	}

	update.Refresh()
	update.Apply()
	update.Commit()

	if recover && update.FailedNamespaces.Contains("0") {
		err := RecoverNode()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("iptables: Failed to recover node iptables")
		}
	}
}

func RemoveNamespace(namespace string) {
	lockId := stateLock.Lock()
	defer stateLock.Unlock(lockId)

	delete(curFailures, namespace)

	if curState == nil {
		return
	}

	for key, rules := range curState.Interfaces {
		if rules.Namespace == namespace {
			delete(curState.Interfaces, key)
		}
	}
}

func UpdateState(nodeSelf *node.Node, vpcs []*vpc.Vpc,
	instances []*instance.Instance, namespaces []string,
	nodeFirewall []*firewall.Rule, firewalls map[string][]*firewall.Rule,
	firewallMaps map[string][]*firewall.Mapping) {

	newState := LoadState(nodeSelf, vpcs, instances, namespaces,
		nodeFirewall, firewalls, firewallMaps)

	ApplyUpdate(newState, namespaces, false)
}

func UpdateStateRecover(nodeSelf *node.Node, vpcs []*vpc.Vpc,
	instances []*instance.Instance, namespaces []string,
	nodeFirewall []*firewall.Rule, firewalls map[string][]*firewall.Rule,
	firewallMaps map[string][]*firewall.Mapping) {

	newState := LoadState(nodeSelf, vpcs, instances, namespaces,
		nodeFirewall, firewalls, firewallMaps)

	ApplyUpdate(newState, namespaces, true)
}

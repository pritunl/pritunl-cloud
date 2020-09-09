package ipset

import (
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/pritunl-cloud/firewall"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
)

var (
	curState      *State
	curNamesState *NamesState
	stateLock     = utils.NewTimeoutLock(3 * time.Minute)
)

func UpdateState(instances []*instance.Instance, namespaces []string,
	nodeFirewall []*firewall.Rule, firewalls map[string][]*firewall.Rule) (
	err error) {

	lockId := stateLock.Lock()
	defer stateLock.Unlock(lockId)

	newState := &State{
		Namespaces: map[string]*Sets{},
	}

	if nodeFirewall != nil {
		newState.AddIngress("0", nodeFirewall)
	}

	for _, inst := range instances {
		if !inst.IsActive() {
			continue
		}

		for i := range inst.Virt.NetworkAdapters {
			namespace := vm.GetNamespace(inst.Id, i)

			ingress := firewalls[namespace]
			if ingress == nil {
				logrus.WithFields(logrus.Fields{
					"instance_id": inst.Id.Hex(),
					"namespace":   namespace,
				}).Warn("ipset: Failed to load instance firewall rules")
				continue
			}

			newState.AddIngress(namespace, ingress)
		}
	}

	err = applyState(curState, newState, namespaces)
	if err != nil {
		return
	}

	curState = newState

	return
}

func applyState(oldState, newState *State, namespaces []string) (err error) {
	namespacesSet := set.NewSet()
	for _, namespace := range namespaces {
		namespacesSet.Add(namespace)
	}

	for _, ipSet := range newState.Namespaces {
		if ipSet.Namespace != "0" && !namespacesSet.Contains(
			ipSet.Namespace) {

			_, err = utils.ExecCombinedOutputLogged(
				[]string{"File exists"},
				"ip", "netns",
				"add", ipSet.Namespace,
			)
			if err != nil {
				return
			}
		}

		curSet := oldState.Namespaces[ipSet.Namespace]

		err = ipSet.Apply(curSet)
		if err != nil {
			return
		}
	}

	return
}

func UpdateNamesState(instances []*instance.Instance,
	nodeFirewall []*firewall.Rule, firewalls map[string][]*firewall.Rule) (
	err error) {

	lockId := stateLock.Lock()
	defer stateLock.Unlock(lockId)

	newNamesState := &NamesState{
		Namespaces: map[string]*Names{},
	}

	if nodeFirewall != nil {
		newNamesState.AddIngress("0", nodeFirewall)
	}

	for _, inst := range instances {
		if !inst.IsActive() {
			continue
		}

		for i := range inst.Virt.NetworkAdapters {
			namespace := vm.GetNamespace(inst.Id, i)

			ingress := firewalls[namespace]
			if ingress == nil {
				logrus.WithFields(logrus.Fields{
					"instance_id": inst.Id.Hex(),
					"namespace":   namespace,
				}).Warn("ipset: Failed to load instance firewall rules")
				continue
			}

			newNamesState.AddIngress(namespace, ingress)
		}
	}

	err = applyNamesState(curNamesState, newNamesState)
	if err != nil {
		return
	}

	curNamesState = newNamesState

	return
}

func applyNamesState(oldNamesState, newNamesState *NamesState) (err error) {
	for _, ipSet := range newNamesState.Namespaces {
		curSet := oldNamesState.Namespaces[ipSet.Namespace]

		err = ipSet.Apply(curSet)
		if err != nil {
			return
		}
	}

	return
}

func loadIpset(namespace string, state *State, namesState *NamesState) (
	err error) {

	output := ""
	if namespace == "0" {
		output, err = utils.ExecOutput("", "ipset", "list")
		if err != nil {
			return
		}
	} else {
		output, err = utils.ExecOutput("",
			"ip", "netns", "exec", namespace, "ipset", "list")
		if err != nil {
			return
		}
	}

	curName := ""
	isMembers := false
	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, "Name:") {
			curName = strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
			isMembers = false
		} else if isMembers {
			if line == "" {
				isMembers = false
			} else {
				member := strings.TrimSpace(line)
				state.AddMember(namespace, curName, member)
				namesState.AddName(namespace, curName)
			}
		} else if strings.HasPrefix(line, "Members:") {
			isMembers = true
		}
	}

	return
}

func Init(namespaces []string, instances []*instance.Instance,
	nodeFirewall []*firewall.Rule, firewalls map[string][]*firewall.Rule) (
	err error) {

	state := &State{
		Namespaces: map[string]*Sets{},
	}
	namesState := &NamesState{
		Namespaces: map[string]*Names{},
	}

	err = loadIpset("0", state, namesState)
	if err != nil {
		return
	}

	for _, namespace := range namespaces {
		err = loadIpset(namespace, state, namesState)
		if err != nil {
			return
		}
	}

	curState = state
	curNamesState = namesState

	err = UpdateState(instances, namespaces, nodeFirewall, firewalls)
	if err != nil {
		return
	}

	return
}

func InitNames(namespaces []string, instances []*instance.Instance,
	nodeFirewall []*firewall.Rule, firewalls map[string][]*firewall.Rule) (
	err error) {

	err = UpdateNamesState(instances, nodeFirewall, firewalls)
	if err != nil {
		return
	}

	return
}

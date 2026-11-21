package ipset

import (
	"strings"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/pritunl-cloud/firewall"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/sirupsen/logrus"
)

var (
	curState      *State
	curNamesState *NamesState
	stateLock     = utils.NewTimeoutLock(3 * time.Minute)
)

func RemoveNamespace(namespace string) {
	lockId := stateLock.Lock()
	defer stateLock.Unlock(lockId)

	if curState != nil {
		delete(curState.Namespaces, namespace)
	}
	if curNamesState != nil {
		delete(curNamesState.Namespaces, namespace)
	}
}

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

	namespacesSet := set.NewSet()
	for _, namespace := range namespaces {
		namespacesSet.Add(namespace)
	}

	for _, inst := range instances {
		for i := range inst.Virt.NetworkAdapters {
			if !inst.IsAvailable(namespacesSet, i) {
				continue
			}

			namespace := vm.GetNamespace(inst.Id, i)
			addr6 := ""
			if len(inst.PrivateIps6) != 0 {
				addr6 = inst.PrivateIps6[0]
			}

			if inst.IsActive() {
				ingress := firewalls[namespace]
				if ingress == nil {
					logrus.WithFields(logrus.Fields{
						"instance_id": inst.Id.Hex(),
						"namespace":   namespace,
					}).Warn("ipset: Failed to load instance firewall rules")
				}
				newState.AddIngress(namespace, ingress)
			}

			newState.AddBase(namespace, addr6)
			if !inst.SkipSourceDestCheck {
				newState.AddSourceDestCheck(namespace, addr6)
			}
		}
	}

	err = applyState(curState, newState, namespacesSet)
	if err != nil {
		return
	}

	curState = newState

	return
}

func applyState(oldState, newState *State, namespacesSet set.Set) (err error) {
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

func UpdateNamesState(instances []*instance.Instance, namespaces []string,
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

	namespacesSet := set.NewSet()
	for _, namespace := range namespaces {
		namespacesSet.Add(namespace)
	}

	for _, inst := range instances {
		for i := range inst.Virt.NetworkAdapters {
			if !inst.IsAvailable(namespacesSet, i) {
				continue
			}

			namespace := vm.GetNamespace(inst.Id, i)
			if inst.IsActive() {
				ingress := firewalls[namespace]
				if ingress == nil {
					logrus.WithFields(logrus.Fields{
						"instance_id": inst.Id.Hex(),
						"namespace":   namespace,
					}).Warn("ipset: Failed to load instance firewall rules")
				}
				newNamesState.AddIngress(namespace, ingress)
			}

			newNamesState.AddBase(namespace)
			if !inst.SkipSourceDestCheck {
				newNamesState.AddSourceDestCheck(namespace)
			}
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

	err = UpdateNamesState(instances, namespaces, nodeFirewall, firewalls)
	if err != nil {
		return
	}

	return
}

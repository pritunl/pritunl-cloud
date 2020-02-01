package iptables

import (
	"fmt"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/firewall"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
)

func diffCmd(a, b []string) bool {
	if len(a) != len(b) {
		return true
	}

	for i := range a {
		if a[i] != b[i] {
			return true
		}
	}

	return false
}

func diffRules(a, b *Rules) bool {
	if len(a.Ingress) != len(b.Ingress) ||
		len(a.Ingress6) != len(b.Ingress6) ||
		len(a.Holds) != len(b.Holds) ||
		len(a.Holds6) != len(b.Holds6) {

		return true
	}

	for i := range a.Ingress {
		if diffCmd(a.Ingress[i], b.Ingress[i]) {
			return true
		}
	}
	for i := range a.Ingress6 {
		if diffCmd(a.Ingress6[i], b.Ingress6[i]) {
			return true
		}
	}
	for i := range a.Holds {
		if diffCmd(a.Holds[i], b.Holds[i]) {
			return true
		}
	}
	for i := range a.Holds6 {
		if diffCmd(a.Holds6[i], b.Holds6[i]) {
			return true
		}
	}

	return false
}

func diffRulesNat(a, b *Rules) bool {
	if a.Nat != b.Nat ||
		a.NatAddr != b.NatAddr ||
		a.NatPubAddr != b.NatPubAddr ||
		a.Nat6 != b.Nat6 ||
		a.NatAddr6 != b.NatAddr6 ||
		a.NatPubAddr6 != b.NatPubAddr6 {

		return true
	}

	return false
}

func getIptablesCmd(ipv6 bool) string {
	if ipv6 {
		return "ip6tables"
	} else {
		return "iptables"
	}
}

func loadIptablesNat(state *State) (err error) {
	Lock()
	defer Unlock()

	hostNat := false
	hostNatInterface := ""
	hostNatExcludes := set.NewSet()
	iptablesCmd := getIptablesCmd(false)

	output, err := utils.ExecOutput("", iptablesCmd, "-t", "nat", "-S")
	if err != nil {
		return
	}

	for _, line := range strings.Split(output, "\n") {
		if !strings.Contains(line, "POSTROUTING") ||
			!strings.Contains(line, "pritunl_cloud_host_nat") {

			continue
		}

		cmd := strings.Fields(line)
		cmdLen := len(cmd)
		if cmdLen < 3 {
			logrus.WithFields(logrus.Fields{
				"iptables_rule": line,
			}).Error("iptables: Invalid iptables state")

			err = &errortypes.ParseError{
				errors.New("iptables: Invalid iptables state"),
			}
			return
		}

		switch cmd[cmdLen-1] {
		case "ACCEPT":
			found := false
			for _, field := range cmd {
				if found {
					hostNatExcludes.Add(field)
					break
				} else if field == "-d" {
					found = true
				}
			}
			break
		case "MASQUERADE":
			found := false
			for _, field := range cmd {
				if found {
					hostNat = true
					hostNatInterface = field
					break
				} else if field == "-o" {
					found = true
				}
			}
			break
		default:
			// TODO Remove invalid rule
		}
	}

	state.HostNat = hostNat
	state.HostNatInterface = hostNatInterface
	state.HostNatExcludes = hostNatExcludes

	return
}

func loadIptables(namespace string, state *State, ipv6 bool) (err error) {
	Lock()
	defer Unlock()

	iptablesCmd := getIptablesCmd(ipv6)

	output := ""
	if namespace == "0" {
		output, err = utils.ExecOutput("", iptablesCmd, "-S")
		if err != nil {
			return
		}
	} else {
		output, err = utils.ExecOutput("",
			"ip", "netns", "exec", namespace, iptablesCmd, "-S")
		if err != nil {
			return
		}
	}

	for _, line := range strings.Split(output, "\n") {
		if !strings.Contains(line, "pritunl_cloud_rule") &&
			!strings.Contains(line, "pritunl_cloud_hold") {

			continue
		}

		cmd := strings.Fields(line)
		if len(cmd) < 3 {
			logrus.WithFields(logrus.Fields{
				"iptables_rule": line,
			}).Error("iptables: Invalid iptables state")

			err = &errortypes.ParseError{
				errors.New("iptables: Invalid iptables state"),
			}
			return
		}
		cmd = cmd[1:]

		iface := ""
		if namespace != "0" {
			if cmd[0] != "FORWARD" {
				logrus.WithFields(logrus.Fields{
					"iptables_rule": line,
				}).Error("iptables: Invalid iptables chain")

				err = &errortypes.ParseError{
					errors.New("iptables: Invalid iptables chain"),
				}
				return
			}

			for i, item := range cmd {
				if item == "--physdev-out" || item == "-o" || item == "-i" {
					if len(cmd) < i+2 {
						logrus.WithFields(logrus.Fields{
							"iptables_rule": line,
						}).Error("iptables: Invalid iptables interface")

						err = &errortypes.ParseError{
							errors.New("iptables: Invalid iptables interface"),
						}
						return
					}
					iface = cmd[i+1]
					break
				}
			}
		} else {
			iface = "host"

			if cmd[0] != "INPUT" {
				logrus.WithFields(logrus.Fields{
					"iptables_rule": line,
				}).Error("iptables: Invalid iptables chain")

				err = &errortypes.ParseError{
					errors.New("iptables: Invalid iptables chain"),
				}
				return
			}
		}

		if iface == "" {
			logrus.WithFields(logrus.Fields{
				"iptables_rule": line,
			}).Error("iptables: Missing iptables interface")

			err = &errortypes.ParseError{
				errors.New("iptables: Missing iptables interface"),
			}
			return
		}

		rules := state.Interfaces[namespace+"-"+iface]
		if rules == nil {
			rules = &Rules{
				Namespace: namespace,
				Interface: iface,
				Ingress:   [][]string{},
				Ingress6:  [][]string{},
				Holds:     [][]string{},
				Holds6:    [][]string{},
			}
			state.Interfaces[namespace+"-"+iface] = rules
		}

		if strings.Contains(line, "pritunl_cloud_hold") {
			if ipv6 {
				rules.Holds6 = append(rules.Holds6, cmd)
			} else {
				rules.Holds = append(rules.Holds, cmd)
			}
		} else {
			if ipv6 {
				rules.Ingress6 = append(rules.Ingress6, cmd)
			} else {
				rules.Ingress = append(rules.Ingress, cmd)
			}
		}
	}

	return
}

func applyState(oldState, newState *State, namespaces []string) (err error) {
	changed := false
	oldIfaces := set.NewSet()
	newIfaces := set.NewSet()

	namespacesSet := set.NewSet()
	for _, namespace := range namespaces {
		namespacesSet.Add(namespace)
	}

	for iface := range oldState.Interfaces {
		oldIfaces.Add(iface)
	}
	for iface := range newState.Interfaces {
		newIfaces.Add(iface)
	}

	oldIfaces.Subtract(newIfaces)
	for iface := range oldIfaces.Iter() {
		err = oldState.Interfaces[iface.(string)].Remove()
		if err != nil {
			return
		}
	}

	iptablesCmd := getIptablesCmd(false)
	if oldState.HostNat != newState.HostNat ||
		oldState.HostNatInterface != newState.HostNatInterface {

		if newState.HostNat {
			if oldState.HostNat {
				_, err = utils.ExecCombinedOutputLogged(
					[]string{
						"matching rule exist",
						"match by that name",
					},
					iptablesCmd,
					"-t", "nat",
					"-D", "POSTROUTING",
					"-o", oldState.HostNatInterface,
					"-m", "comment",
					"--comment", "pritunl_cloud_host_nat",
					"-j", "MASQUERADE",
				)
				if err != nil {
					return
				}
			}
			_, err = utils.ExecCombinedOutputLogged(
				[]string{
					"matching rule exist",
				},
				iptablesCmd,
				"-t", "nat",
				"-A", "POSTROUTING",
				"-o", newState.HostNatInterface,
				"-m", "comment",
				"--comment", "pritunl_cloud_host_nat",
				"-j", "MASQUERADE",
			)
			if err != nil {
				return
			}
		} else if oldState.HostNat {
			_, err = utils.ExecCombinedOutputLogged(
				[]string{
					"matching rule exist",
					"match by that name",
				},
				iptablesCmd,
				"-t", "nat",
				"-D", "POSTROUTING",
				"-o", oldState.HostNatInterface,
				"-m", "comment",
				"--comment", "pritunl_cloud_host_nat",
				"-j", "MASQUERADE",
			)
			if err != nil {
				return
			}
		}
	}

	remNatExcludes := oldState.HostNatExcludes.Copy()
	remNatExcludes.Subtract(newState.HostNatExcludes)
	for natExcludeInf := range remNatExcludes.Iter() {
		natExclude := natExcludeInf.(string)
		_, err = utils.ExecCombinedOutputLogged(
			[]string{
				"matching rule exist",
			},
			iptablesCmd,
			"-t", "nat",
			"-D", "POSTROUTING",
			"-d", natExclude,
			"-m", "comment",
			"--comment", "pritunl_cloud_host_nat",
			"-j", "ACCEPT",
		)
		if err != nil {
			return
		}
	}

	addNatExcludes := newState.HostNatExcludes.Copy()
	addNatExcludes.Subtract(oldState.HostNatExcludes)
	for natExcludeInf := range addNatExcludes.Iter() {
		natExclude := natExcludeInf.(string)
		_, err = utils.ExecCombinedOutputLogged(
			[]string{
				"matching rule exist",
			},
			iptablesCmd,
			"-t", "nat",
			"-I", "POSTROUTING", "1",
			"-d", natExclude,
			"-m", "comment",
			"--comment", "pritunl_cloud_host_nat",
			"-j", "ACCEPT",
		)
		if err != nil {
			return
		}
	}

	for _, rules := range newState.Interfaces {
		if rules.Namespace != "0" && !namespacesSet.Contains(rules.Namespace) {
			_, err = utils.ExecCombinedOutputLogged(
				[]string{"File exists"},
				"ip", "netns",
				"add", rules.Namespace,
			)
			if err != nil {
				return
			}
		}

		oldRules := oldState.Interfaces[rules.Namespace+"-"+rules.Interface]
		if oldRules != nil {
			if !diffRules(oldRules, rules) {
				continue
			}

			if !changed {
				changed = true
				logrus.Info("iptables: Updating iptables")
			}

			if rules.Interface != "host" {
				err = rules.Hold()
				if err != nil {
					return
				}
			}

			err = oldRules.Remove()
			if err != nil {
				return
			}
		}

		if !changed {
			changed = true
			logrus.Info("iptables: Updating iptables")
		}

		err = rules.Apply()
		if err != nil {
			return
		}
	}

	return
}

func UpdateState(nodeSelf *node.Node, instances []*instance.Instance,
	namespaces []string, nodeFirewall []*firewall.Rule,
	firewalls map[string][]*firewall.Rule) (err error) {

	lockId := stateLock.Lock()
	defer stateLock.Unlock(lockId)

	nodeNetworkMode := node.Self.NetworkMode
	if nodeNetworkMode == "" {
		nodeNetworkMode = node.Dhcp
	}
	nodeNetworkMode6 := node.Self.NetworkMode6

	externalNetwork := true
	if nodeNetworkMode == node.Internal {
		externalNetwork = false
	}

	externalNetwork6 := false
	if nodeNetworkMode6 != "" && (nodeNetworkMode != nodeNetworkMode6 ||
		(nodeNetworkMode6 == node.Static)) {

		externalNetwork6 = true
	}

	newState := &State{
		Interfaces: map[string]*Rules{},
	}

	if nodeFirewall != nil {
		newState.Interfaces["0-host"] = generate("0", "host", nodeFirewall)
	}

	hostNat := false
	hostNetwork := false
	natExcludesSet := set.NewSet()
	if !nodeSelf.HostBlock.IsZero() && nodeSelf.DefaultInterface != "" {
		hostNetwork = true
		hostNat = nodeSelf.HostNat
		natExcludes := nodeSelf.HostNatExcludes
		if hostNat && natExcludes != nil {
			for _, natExclude := range natExcludes {
				natExcludesSet.Add(natExclude)
			}
		}
		newState.HostNatInterface = nodeSelf.DefaultInterface
	}
	newState.HostNat = hostNat
	newState.HostNatExcludes = natExcludesSet

	for _, inst := range instances {
		if !inst.IsActive() {
			continue
		}

		namespace := vm.GetNamespace(inst.Id, 0)
		iface := vm.GetIface(inst.Id, 0)
		ifaceExternal := vm.GetIfaceExternal(inst.Id, 0)
		ifaceExternal6 := vm.GetIfaceExternal(inst.Id, 1)
		ifaceHost := vm.GetIfaceHost(inst.Id, 0)

		_, ok := newState.Interfaces[namespace+"-"+iface]
		if ok {
			logrus.WithFields(logrus.Fields{
				"namespace": namespace,
				"interface": iface,
			}).Error("iptables: Virtual interface conflict")

			err = &errortypes.ParseError{
				errors.New("iptables: Virtual interface conflict"),
			}
			return
		}

		ingress := firewalls[namespace]
		if ingress == nil {
			logrus.WithFields(logrus.Fields{
				"instance_id": inst.Id.Hex(),
				"namespace":   namespace,
			}).Warn("iptables: Failed to load instance firewall rules")
			continue
		}

		if externalNetwork {
			rules := generateInternal(namespace, ifaceExternal, ingress)
			newState.Interfaces[namespace+"-"+ifaceExternal] = rules
		}

		if externalNetwork6 {
			rules := generateInternal(namespace, ifaceExternal6, ingress)
			newState.Interfaces[namespace+"-"+ifaceExternal6] = rules
		}

		if hostNetwork {
			rules := generateInternal(namespace, ifaceHost, ingress)
			newState.Interfaces[namespace+"-"+ifaceHost] = rules
		}

		rules := generateVirt(namespace, iface, ingress)
		newState.Interfaces[namespace+"-"+iface] = rules
	}

	err = applyState(curState, newState, namespaces)
	if err != nil {
		return
	}

	curState = newState

	return
}

func Recover() (err error) {
	cmds := [][]string{}

	if !node.Self.Firewall {
		return
	}

	cmds = append(cmds, []string{
		"-I", "INPUT", "1",
		"-m", "comment",
		"--comment", "pritunl_cloud_rule",
		"-j", "DROP",
	})
	cmds = append(cmds, []string{
		"-I", "INPUT", "1",
		"-m", "conntrack",
		"--ctstate", "INVALID",
		"-m", "comment",
		"--comment", "pritunl_cloud_rule",
		"-j", "DROP",
	})
	cmds = append(cmds, []string{
		"-I", "INPUT", "1",
		"-m", "conntrack",
		"--ctstate", "RELATED,ESTABLISHED",
		"-m", "comment", "--comment", "pritunl_cloud_rule",
		"-j", "ACCEPT",
	})

	for _, cmd := range cmds {
		Lock()
		output, e := utils.ExecCombinedOutput("", "iptables", cmd...)
		Unlock()
		if e != nil {
			err = e
			logrus.WithFields(logrus.Fields{
				"command": cmd,
				"output":  output,
				"error":   err,
			}).Error("iptables: Failed to add iptables recover rule")
			return
		}
	}

	for _, cmd := range cmds {
		Lock()
		output, e := utils.ExecCombinedOutput("", "ip6tables", cmd...)
		Unlock()
		if e != nil {
			err = e
			logrus.WithFields(logrus.Fields{
				"command": cmd,
				"output":  output,
				"error":   err,
			}).Error("iptables: Failed to add ip6tables recover rule")
			return
		}
	}

	time.Sleep(10 * time.Second)

	db := database.GetDatabase()
	defer db.Close()

	namespaces, err := utils.GetNamespaces()
	if err != nil {
		return
	}

	disks, err := disk.GetNode(db, node.Self.Id)
	if err != nil {
		return
	}

	instances, err := instance.GetAllVirt(db, &bson.M{
		"node": node.Self.Id,
	}, disks)
	if err != nil {
		return
	}

	nodeFirewall, firewalls, err := firewall.GetAllIngress(
		db, node.Self, instances)
	if err != nil {
		return
	}

	err = Init(namespaces, instances, nodeFirewall, firewalls)
	if err != nil {
		return
	}

	return
}

func Init(namespaces []string, instances []*instance.Instance,
	nodeFirewall []*firewall.Rule, firewalls map[string][]*firewall.Rule) (
	err error) {

	_, err = utils.ExecCombinedOutputLogged(
		nil, "sysctl", "-w", "net.ipv6.conf.all.accept_ra=2",
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil, "sysctl", "-w", "net.ipv6.conf.default.accept_ra=2",
	)
	if err != nil {
		return
	}

	interfaces, err := utils.GetInterfaces()
	if err != nil {
		return
	}

	for _, iface := range interfaces {
		if len(iface) == 14 && (strings.HasPrefix(iface, "v") ||
			strings.HasPrefix(iface, "x")) {

			continue
		}

		utils.ExecCombinedOutput("",
			"sysctl", "-w",
			fmt.Sprintf("net.ipv6.conf.%s.accept_ra=2", iface),
		)
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil, "sysctl", "-w", "net.ipv4.ip_forward=1",
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil, "sysctl", "-w", "net.ipv6.conf.all.forwarding=1",
	)
	if err != nil {
		return
	}

	state := &State{
		Interfaces: map[string]*Rules{},
	}

	err = loadIptablesNat(state)
	if err != nil {
		return
	}

	err = loadIptables("0", state, false)
	if err != nil {
		return
	}

	err = loadIptables("0", state, true)
	if err != nil {
		return
	}

	for _, namespace := range namespaces {
		err = loadIptables(namespace, state, false)
		if err != nil {
			return
		}

		err = loadIptables(namespace, state, true)
		if err != nil {
			return
		}
	}

	curState = state

	err = UpdateState(node.Self, instances, namespaces,
		nodeFirewall, firewalls)
	if err != nil {
		return
	}

	return
}

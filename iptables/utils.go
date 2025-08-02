package iptables

import (
	"fmt"
	"strings"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/firewall"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/pritunl/pritunl-cloud/vpc"
	"github.com/sirupsen/logrus"
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

func diffRules(a, b *Rules) *RulesDiff {
	diff := &RulesDiff{}
	changed := false

	if len(a.Header) != len(b.Header) {
		diff.HeaderDiff = true
		changed = true
	} else {
		for i := range a.Header {
			if diffCmd(a.Header[i], b.Header[i]) {
				diff.HeaderDiff = true
				changed = true
				break
			}
		}
	}

	if len(a.Header6) != len(b.Header6) {
		diff.Header6Diff = true
		changed = true
	} else {
		for i := range a.Header6 {
			if diffCmd(a.Header6[i], b.Header6[i]) {
				diff.Header6Diff = true
				changed = true
				break
			}
		}
	}

	if len(a.SourceDestCheck) != len(b.SourceDestCheck) {
		diff.SourceDestCheckDiff = true
		changed = true
	} else {
		for i := range a.SourceDestCheck {
			if diffCmd(a.SourceDestCheck[i], b.SourceDestCheck[i]) {
				diff.SourceDestCheckDiff = true
				changed = true
				break
			}
		}
	}

	if len(a.SourceDestCheck6) != len(b.SourceDestCheck6) {
		diff.SourceDestCheck6Diff = true
		changed = true
	} else {
		for i := range a.SourceDestCheck6 {
			if diffCmd(a.SourceDestCheck6[i], b.SourceDestCheck6[i]) {
				diff.SourceDestCheck6Diff = true
				changed = true
				break
			}
		}
	}

	if len(a.Ingress) != len(b.Ingress) {
		diff.IngressDiff = true
		changed = true
	} else {
		for i := range a.Ingress {
			if diffCmd(a.Ingress[i], b.Ingress[i]) {
				diff.IngressDiff = true
				changed = true
				break
			}
		}
	}

	if len(a.Ingress6) != len(b.Ingress6) {
		diff.Ingress6Diff = true
		changed = true
	} else {
		for i := range a.Ingress6 {
			if diffCmd(a.Ingress6[i], b.Ingress6[i]) {
				diff.Ingress6Diff = true
				changed = true
				break
			}
		}
	}

	if len(a.Nats) != len(b.Nats) {
		diff.NatsDiff = true
		changed = true
	} else {
		for i := range a.Nats {
			if diffCmd(a.Nats[i], b.Nats[i]) {
				diff.NatsDiff = true
				changed = true
				break
			}
		}
	}

	if len(a.Nats6) != len(b.Nats6) {
		diff.Nats6Diff = true
		changed = true
	} else {
		for i := range a.Nats6 {
			if diffCmd(a.Nats6[i], b.Nats6[i]) {
				diff.Nats6Diff = true
				changed = true
				break
			}
		}
	}

	if len(a.Maps) != len(b.Maps) {
		diff.MapsDiff = true
		changed = true
	} else {
		for i := range a.Maps {
			if diffCmd(a.Maps[i], b.Maps[i]) {
				diff.MapsDiff = true
				changed = true
				break
			}
		}
	}

	if len(a.Maps6) != len(b.Maps6) {
		diff.Maps6Diff = true
		changed = true
	} else {
		for i := range a.Maps6 {
			if diffCmd(a.Maps6[i], b.Maps6[i]) {
				diff.Maps6Diff = true
				changed = true
				break
			}
		}
	}

	if len(a.Holds) != len(b.Holds) {
		diff.HoldsDiff = true
		changed = true
	} else {
		for i := range a.Holds {
			if diffCmd(a.Holds[i], b.Holds[i]) {
				diff.HoldsDiff = true
				changed = true
				break
			}
		}
	}

	if len(a.Holds6) != len(b.Holds6) {
		diff.Holds6Diff = true
		changed = true
	} else {
		for i := range a.Holds6 {
			if diffCmd(a.Holds6[i], b.Holds6[i]) {
				diff.Holds6Diff = true
				changed = true
				break
			}
		}
	}

	if !changed {
		return nil
	}

	return diff
}

func getIptablesCmd(ipv6 bool) string {
	if ipv6 {
		return "ip6tables"
	} else {
		return "iptables"
	}
}

func loadIptables(namespace, instIface string, state *State,
	ipv6 bool) (err error) {

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
		ruleComment := strings.Contains(line, "pritunl_cloud_rule")
		holdComment := strings.Contains(line, "pritunl_cloud_hold")
		headComment := strings.Contains(line, "pritunl_cloud_head")
		sdcComment := strings.Contains(line, "pritunl_cloud_sdc")

		if !ruleComment && !holdComment && !headComment && !sdcComment {
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
		if sdcComment {
			if cmd[0] != "FORWARD" {
				logrus.WithFields(logrus.Fields{
					"iptables_rule": line,
				}).Error("iptables: Invalid iptables sdc chain")

				err = &errortypes.ParseError{
					errors.New("iptables: Invalid iptables sdc chain"),
				}
				return
			}

			for i, item := range cmd {
				if item == "--physdev-in" || item == "--physdev-out" {
					if len(cmd) < i+2 {
						logrus.WithFields(logrus.Fields{
							"iptables_rule": line,
						}).Error("iptables: Invalid iptables sdc interface")

						err = &errortypes.ParseError{
							errors.New(
								"iptables: Invalid iptables sdc interface"),
						}
						return
					}
					iface = strings.Trim(cmd[i+1], "+")
					break
				}
			}
		} else if namespace != "0" {
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
							errors.New(
								"iptables: Invalid iptables interface"),
						}
						return
					}
					iface = strings.Trim(cmd[i+1], "+")
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
				Namespace:        namespace,
				Interface:        iface,
				Header:           [][]string{},
				Header6:          [][]string{},
				SourceDestCheck:  [][]string{},
				SourceDestCheck6: [][]string{},
				Ingress:          [][]string{},
				Ingress6:         [][]string{},
				Maps:             [][]string{},
				Maps6:            [][]string{},
				Holds:            [][]string{},
				Holds6:           [][]string{},
			}
			state.Interfaces[namespace+"-"+iface] = rules
		}

		if holdComment {
			if ipv6 {
				rules.Holds6 = append(rules.Holds6, cmd)
			} else {
				rules.Holds = append(rules.Holds, cmd)
			}
		} else if sdcComment {
			if ipv6 {
				rules.SourceDestCheck6 = append(rules.SourceDestCheck6, cmd)
			} else {
				rules.SourceDestCheck = append(rules.SourceDestCheck, cmd)
			}
		} else {
			if headComment {
				if ipv6 {
					rules.Header6 = append(rules.Header6, cmd)
				} else {
					rules.Header = append(rules.Header, cmd)
				}
			} else {
				if ipv6 {
					rules.Ingress6 = append(rules.Ingress6, cmd)
				} else {
					rules.Ingress = append(rules.Ingress, cmd)
				}
			}
		}
	}

	if namespace == "0" {
		output, err = utils.ExecOutput("", iptablesCmd, "-S", "-t", "nat")
		if err != nil {
			return
		}
	} else {
		output, err = utils.ExecOutput("",
			"ip", "netns", "exec", namespace, iptablesCmd, "-S", "-t", "nat")
		if err != nil {
			return
		}
	}

	postIface := ""
	natRules := [][]string{}

	for _, line := range strings.Split(output, "\n") {
		natComment := strings.Contains(line, "pritunl_cloud_nat")
		mapComment := strings.Contains(line, "pritunl_cloud_map")

		if !natComment && !mapComment {
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

		if mapComment && namespace == "0" {
			if cmd[0] != "PREROUTING" && cmd[0] != "POSTROUTING" {
				logrus.WithFields(logrus.Fields{
					"iptables_rule": line,
				}).Error("iptables: Invalid iptables map chain")

				err = &errortypes.ParseError{
					errors.New("iptables: Invalid iptables map chain"),
				}
				return
			}

			rules := state.Interfaces[namespace+"-host"]
			if rules == nil {
				rules = &Rules{
					Namespace:        namespace,
					Interface:        "host",
					Header:           [][]string{},
					Header6:          [][]string{},
					SourceDestCheck:  [][]string{},
					SourceDestCheck6: [][]string{},
					Ingress:          [][]string{},
					Ingress6:         [][]string{},
					Maps:             [][]string{},
					Maps6:            [][]string{},
					Holds:            [][]string{},
					Holds6:           [][]string{},
				}
				state.Interfaces[namespace+"-host"] = rules
			}

			if ipv6 {
				rules.Maps6 = append(rules.Maps6, cmd)
			} else {
				rules.Maps = append(rules.Maps, cmd)
			}
		} else if mapComment {
			iface := instIface

			for i, item := range cmd {
				if item == "-i" {
					if len(cmd) < i+2 {
						logrus.WithFields(logrus.Fields{
							"iptables_rule": line,
						}).Error("iptables: Invalid iptables interface")

						err = &errortypes.ParseError{
							errors.New(
								"iptables: Invalid iptables interface"),
						}
						return
					}
					iface = strings.Trim(cmd[i+1], "+")
					break
				}
			}

			if iface == "" {
				logrus.WithFields(logrus.Fields{
					"namespace":     namespace,
					"iface":         iface,
					"iptables_rule": line,
				}).Error("iptables: Missing instance iface for map")
			} else {
				if cmd[0] != "PREROUTING" {
					logrus.WithFields(logrus.Fields{
						"iptables_rule": line,
					}).Error("iptables: Invalid iptables map chain")

					err = &errortypes.ParseError{
						errors.New("iptables: Invalid iptables map chain"),
					}
					return
				}

				rules := state.Interfaces[namespace+"-"+iface]
				if rules == nil {
					rules = &Rules{
						Namespace:        namespace,
						Interface:        iface,
						Header:           [][]string{},
						Header6:          [][]string{},
						SourceDestCheck:  [][]string{},
						SourceDestCheck6: [][]string{},
						Ingress:          [][]string{},
						Ingress6:         [][]string{},
						Maps:             [][]string{},
						Maps6:            [][]string{},
						Holds:            [][]string{},
						Holds6:           [][]string{},
					}
					state.Interfaces[namespace+"-"+iface] = rules
				}

				if ipv6 {
					rules.Maps6 = append(rules.Maps6, cmd)
				} else {
					rules.Maps = append(rules.Maps, cmd)
				}
			}
		} else if natComment {
			if cmd[0] != "PREROUTING" && cmd[0] != "POSTROUTING" {
				logrus.WithFields(logrus.Fields{
					"iptables_rule": line,
				}).Error("iptables: Invalid iptables map chain")

				err = &errortypes.ParseError{
					errors.New("iptables: Invalid iptables map chain"),
				}
				return
			}

			if cmd[0] == "POSTROUTING" {
				for i, item := range cmd {
					if item == "-o" {
						if len(cmd) < i+2 {
							logrus.WithFields(logrus.Fields{
								"iptables_rule": line,
							}).Error("iptables: Invalid iptables addr")

							err = &errortypes.ParseError{
								errors.New(
									"iptables: Invalid iptables addr"),
							}
							return
						}
						postIface = strings.Trim(cmd[i+1], "+")
					}
				}
			}

			natRules = append(natRules, cmd)
		}
	}

	cloudPostIface := ""
	cloudNatRules := [][]string{}

	for _, line := range strings.Split(output, "\n") {
		if !strings.Contains(line, "pritunl_cloud_cloud_nat") {
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

		if cmd[0] != "PREROUTING" && cmd[0] != "POSTROUTING" {
			logrus.WithFields(logrus.Fields{
				"iptables_rule": line,
			}).Error("iptables: Invalid iptables map chain")

			err = &errortypes.ParseError{
				errors.New("iptables: Invalid iptables map chain"),
			}
			return
		}

		if cmd[0] == "POSTROUTING" {
			for i, item := range cmd {
				if item == "-o" {
					if len(cmd) < i+2 {
						logrus.WithFields(logrus.Fields{
							"iptables_rule": line,
						}).Error("iptables: Invalid iptables addr")

						err = &errortypes.ParseError{
							errors.New(
								"iptables: Invalid iptables addr"),
						}
						return
					}
					cloudPostIface = strings.Trim(cmd[i+1], "+")
				}
			}
		}

		cloudNatRules = append(cloudNatRules, cmd)
	}

	if postIface != "" {
		rules := state.Interfaces[namespace+"-"+postIface]
		if rules == nil {
			rules = &Rules{
				Namespace:        namespace,
				Interface:        postIface,
				Header:           [][]string{},
				Header6:          [][]string{},
				SourceDestCheck:  [][]string{},
				SourceDestCheck6: [][]string{},
				Ingress:          [][]string{},
				Ingress6:         [][]string{},
				Holds:            [][]string{},
				Holds6:           [][]string{},
			}
			state.Interfaces[namespace+"-"+postIface] = rules
		}

		if ipv6 {
			rules.Nats6 = append(rules.Nats6, natRules...)
		} else {
			rules.Nats = append(rules.Nats, natRules...)
		}
	}

	if cloudPostIface != "" {
		rules := state.Interfaces[namespace+"-"+cloudPostIface]
		if rules == nil {
			rules = &Rules{
				Namespace:        namespace,
				Interface:        cloudPostIface,
				Header:           [][]string{},
				Header6:          [][]string{},
				SourceDestCheck:  [][]string{},
				SourceDestCheck6: [][]string{},
				Ingress:          [][]string{},
				Ingress6:         [][]string{},
				Holds:            [][]string{},
				Holds6:           [][]string{},
			}
			state.Interfaces[namespace+"-"+cloudPostIface] = rules
		}

		if ipv6 {
			rules.Nats6 = append(rules.Nats6, cloudNatRules...)
		} else {
			rules.Nats = append(rules.Nats, cloudNatRules...)
		}
	}

	return
}

func RecoverNode() (err error) {
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

	return
}

func Init(namespaces []string, vpcs []*vpc.Vpc,
	instances []*instance.Instance, nodeFirewall []*firewall.Rule,
	firewalls map[string][]*firewall.Rule,
	firewallMaps map[string][]*firewall.Mapping) (err error) {

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

	utils.ExecCombinedOutput(
		"", "sysctl", "-w", "net.bridge.bridge-nf-call-iptables=1",
	)
	utils.ExecCombinedOutput(
		"", "sysctl", "-w", "net.bridge.bridge-nf-call-ip6tables=1",
	)

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

	err = loadIptables("0", "", state, false)
	if err != nil {
		return
	}

	err = loadIptables("0", "", state, true)
	if err != nil {
		return
	}

	namespaceMap := map[string]*instance.Instance{}
	for _, inst := range instances {
		namespaceMap[vm.GetNamespace(inst.Id, 0)] = inst
	}

	for _, namespace := range namespaces {
		instIface := ""
		inst := namespaceMap[namespace]
		if inst != nil {
			instIface = vm.GetIface(inst.Id, 0)
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

	curState = state

	UpdateState(node.Self, vpcs, instances,
		namespaces, nodeFirewall, firewalls, firewallMaps)

	return
}

func protocolIndex(proto string) string {
	switch proto {
	case "icmp":
		return "1"
	case "tcp":
		return "6"
	case "udp":
		return "17"
	default:
		return "0"
	}
}

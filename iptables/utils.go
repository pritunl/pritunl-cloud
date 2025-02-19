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

func diffRules(a, b *Rules) bool {
	if len(a.Header) != len(b.Header) ||
		len(a.Header6) != len(b.Header6) ||
		len(a.SourceDestCheck) != len(b.SourceDestCheck) ||
		len(a.SourceDestCheck6) != len(b.SourceDestCheck6) ||
		len(a.Ingress) != len(b.Ingress) ||
		len(a.Ingress6) != len(b.Ingress6) ||
		len(a.Maps) != len(b.Maps) ||
		len(a.Maps6) != len(b.Maps6) ||
		len(a.Holds) != len(b.Holds) ||
		len(a.Holds6) != len(b.Holds6) {

		return true
	}

	for i := range a.Header {
		if diffCmd(a.Header[i], b.Header[i]) {
			return true
		}
	}
	for i := range a.Header6 {
		if diffCmd(a.Header6[i], b.Header6[i]) {
			return true
		}
	}
	for i := range a.SourceDestCheck {
		if diffCmd(a.SourceDestCheck[i], b.SourceDestCheck[i]) {
			return true
		}
	}
	for i := range a.SourceDestCheck6 {
		if diffCmd(a.SourceDestCheck6[i], b.SourceDestCheck6[i]) {
			return true
		}
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
	for i := range a.Maps {
		if diffCmd(a.Maps[i], b.Maps[i]) {
			return true
		}
	}
	for i := range a.Maps6 {
		if diffCmd(a.Maps6[i], b.Maps6[i]) {
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

func diffRulesNat(a, b *Rules) (bool, string) {
	if a.Nat != b.Nat {
		return true, "nat"
	}

	if a.NatAddr != b.NatAddr {
		return true, "nat_addr"
	}

	if a.NatPubAddr != b.NatPubAddr {
		return true, "nat_pub_addr"
	}

	if a.Nat6 != b.Nat6 {
		return true, "nat6"
	}

	if a.NatAddr6 != b.NatAddr6 {
		return true, "nat_addr6"
	}

	if a.NatPubAddr6 != b.NatPubAddr6 {
		return true, "nat_pub_addr6"
	}

	if a.OracleNat != b.OracleNat {
		return true, "oracle_nat"
	}

	if a.OracleNatAddr != b.OracleNatAddr {
		return true, "oracle_nat_addr"
	}

	if a.OracleNatPubAddr != b.OracleNatPubAddr {
		return true, "oracle_nat_pub_addr"
	}

	return false, ""
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
					iface = cmd[i+1]
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

	preAddr := ""
	prePubAddr := ""
	postAddr := ""
	postIface := ""

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
			if cmd[0] != "PREROUTING" {
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
					iface = cmd[i+1]
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
		} else {
			switch cmd[0] {
			case "PREROUTING":
				for i, item := range cmd {
					if item == "-d" {
						if len(cmd) < i+2 {
							logrus.WithFields(logrus.Fields{
								"iptables_rule": line,
							}).Error("iptables: Invalid iptables pub addr")

							err = &errortypes.ParseError{
								errors.New(
									"iptables: Invalid iptables pub addr"),
							}
							return
						}
						prePubAddr = strings.Split(cmd[i+1], "/")[0]
					}

					if item == "--to-destination" {
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
						preAddr = strings.Split(cmd[i+1], "/")[0]
					}
				}
				break
			case "POSTROUTING":
				for i, item := range cmd {
					if item == "-s" {
						if len(cmd) < i+2 {
							logrus.WithFields(logrus.Fields{
								"iptables_rule": line,
							}).Error("iptables: Invalid iptables pub addr")

							err = &errortypes.ParseError{
								errors.New(
									"iptables: Invalid iptables pub addr"),
							}
							return
						}
						postAddr = strings.Split(cmd[i+1], "/")[0]
					}

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
						postIface = cmd[i+1]
					}
				}
				break
			}
		}
	}

	oraclePreAddr := ""
	oraclePrePubAddr := ""
	oraclePostAddr := ""
	oraclePostIface := ""

	for _, line := range strings.Split(output, "\n") {
		if !strings.Contains(line, "pritunl_cloud_oracle_nat") {
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

		switch cmd[0] {
		case "PREROUTING":
			for i, item := range cmd {
				if item == "-d" {
					if len(cmd) < i+2 {
						logrus.WithFields(logrus.Fields{
							"iptables_rule": line,
						}).Error("iptables: Invalid iptables pub addr")

						err = &errortypes.ParseError{
							errors.New(
								"iptables: Invalid iptables pub addr"),
						}
						return
					}
					oraclePrePubAddr = strings.Split(cmd[i+1], "/")[0]
				}

				if item == "--to-destination" {
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
					oraclePreAddr = strings.Split(cmd[i+1], "/")[0]
				}
			}
			break
		case "POSTROUTING":
			for i, item := range cmd {
				if item == "-s" {
					if len(cmd) < i+2 {
						logrus.WithFields(logrus.Fields{
							"iptables_rule": line,
						}).Error("iptables: Invalid iptables pub addr")

						err = &errortypes.ParseError{
							errors.New(
								"iptables: Invalid iptables pub addr"),
						}
						return
					}
					oraclePostAddr = strings.Split(cmd[i+1], "/")[0]
				}

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
					oraclePostIface = cmd[i+1]
				}
			}
			break
		}
	}

	if preAddr != "" && prePubAddr != "" && postIface != "" &&
		postAddr == preAddr {

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

		if !ipv6 {
			rules.Nat = true
			rules.NatAddr = preAddr
			rules.NatPubAddr = prePubAddr
		} else {
			rules.Nat6 = true
			rules.NatAddr6 = preAddr
			rules.NatPubAddr6 = prePubAddr
		}
	}

	if oraclePreAddr != "" && oraclePrePubAddr != "" &&
		oraclePostIface != "" && oraclePostAddr == oraclePreAddr {

		rules := state.Interfaces[namespace+"-"+oraclePostIface]
		if rules == nil {
			rules = &Rules{
				Namespace:        namespace,
				Interface:        oraclePostIface,
				Header:           [][]string{},
				Header6:          [][]string{},
				SourceDestCheck:  [][]string{},
				SourceDestCheck6: [][]string{},
				Ingress:          [][]string{},
				Ingress6:         [][]string{},
				Holds:            [][]string{},
				Holds6:           [][]string{},
			}
			state.Interfaces[namespace+"-"+oraclePostIface] = rules
		}

		if oraclePreAddr != "" && oraclePrePubAddr != "" &&
			oraclePostIface != "" && oraclePostAddr == oraclePreAddr {

			rules.OracleNat = true
			rules.OracleNatAddr = oraclePreAddr
			rules.OracleNatPubAddr = oraclePrePubAddr
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

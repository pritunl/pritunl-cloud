package iptables

import (
	"fmt"
	"strings"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/firewall"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/utils"
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
	if len(a.SourceDestCheck) != len(b.SourceDestCheck) ||
		len(a.SourceDestCheck6) != len(b.SourceDestCheck6) ||
		len(a.Ingress) != len(b.Ingress) ||
		len(a.Ingress6) != len(b.Ingress6) ||
		len(a.Holds) != len(b.Holds) ||
		len(a.Holds6) != len(b.Holds6) {

		return true
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
		a.NatPubAddr6 != b.NatPubAddr6 ||
		a.OracleNat != b.OracleNat ||
		a.OracleNatAddr != b.OracleNatAddr ||
		a.OracleNatPubAddr != b.OracleNatPubAddr {

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
		ruleComment := strings.Contains(line, "pritunl_cloud_rule")
		holdComment := strings.Contains(line, "pritunl_cloud_hold")
		sdcComment := strings.Contains(line, "pritunl_cloud_sdc")

		if !ruleComment && !holdComment && !sdcComment {
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
				SourceDestCheck:  [][]string{},
				SourceDestCheck6: [][]string{},
				Ingress:          [][]string{},
				Ingress6:         [][]string{},
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
			if ipv6 {
				rules.Ingress6 = append(rules.Ingress6, cmd)
			} else {
				rules.Ingress = append(rules.Ingress, cmd)
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
		if !strings.Contains(line, "pritunl_cloud_nat") {
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
				Namespace: namespace,
				Interface: oraclePostIface,
				Ingress:   [][]string{},
				Ingress6:  [][]string{},
				Holds:     [][]string{},
				Holds6:    [][]string{},
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

	UpdateState(node.Self, instances, namespaces, nodeFirewall, firewalls)

	return
}

package iptables

import (
	"github.com/Sirupsen/logrus"
	"github.com/pritunl/pritunl-cloud/firewall"
	"github.com/pritunl/pritunl-cloud/utils"
	"strings"
	"time"
)

var (
	curState  *State
	stateLock = utils.NewTimeoutLock(1 * time.Minute)
)

type Rules struct {
	Namespace string
	Interface string
	Ingress   [][]string
	Ingress6  [][]string
	Holds     [][]string
	Holds6    [][]string
}

type State struct {
	Interfaces map[string]*Rules
}

func (r *Rules) newCommand() (cmd []string) {
	chain := ""
	if r.Interface == "host" {
		chain = "INPUT"
	} else {
		chain = "FORWARD"
	}

	cmd = []string{
		chain,
	}

	return
}

func (r *Rules) commentCommand(inCmd []string, hold bool) (cmd []string) {
	comment := ""
	if hold {
		comment = "pritunl_cloud_hold"
	} else {
		comment = "pritunl_cloud_rule"
	}

	cmd = append(inCmd,
		"-m", "comment",
		"--comment", comment,
	)

	return
}

func (r *Rules) run(cmds [][]string, ipCmd string, ipv6 bool) (err error) {
	iptablesCmd := getIptablesCmd(ipv6)

	for _, cmd := range cmds {
		cmd = append([]string{ipCmd}, cmd...)

		for i := 0; i < 3; i++ {
			output := "'"
			if r.Namespace == "0" {
				output, err = utils.ExecCombinedOutputLogged(
					[]string{"matching rule exist"}, iptablesCmd, cmd...)
			} else {
				cmd = append([]string{
					"netns", "exec", r.Namespace,
					iptablesCmd,
				}, cmd...)

				output, err = utils.ExecCombinedOutputLogged(
					[]string{"matching rule exist"}, "ip", cmd...)
			}

			if err != nil {
				if i < 2 {
					err = nil
					time.Sleep(500 * time.Millisecond)
					continue
				} else if cmd[len(cmd)-1] == "ACCEPT" {
					err = nil
					logrus.WithFields(logrus.Fields{
						"ipv6":    ipv6,
						"command": cmd,
						"output":  output,
					}).Error("iptables: Ignoring invalid iptables command")
				} else {
					logrus.WithFields(logrus.Fields{
						"ipv6":    ipv6,
						"command": cmd,
						"output":  output,
					}).Warn("iptables: Failed to run iptables command")
					return
				}
			}
			break
		}
	}

	return
}

func (r *Rules) Apply() (err error) {
	err = r.run(r.Ingress, "-A", false)
	if err != nil {
		return
	}

	err = r.run(r.Ingress6, "-A", true)
	if err != nil {
		return
	}

	err = r.run(r.Holds, "-D", false)
	if err != nil {
		return
	}
	r.Holds = [][]string{}

	err = r.run(r.Holds6, "-D", true)
	if err != nil {
		return
	}
	r.Holds6 = [][]string{}

	return
}

func (r *Rules) Hold() (err error) {
	cmd := r.newCommand()
	if r.Interface != "host" {
		cmd = append(cmd,
			//"-m", "physdev",
			//"--physdev-out", r.Interface,
			"-o", r.Interface,
		)
	}
	cmd = r.commentCommand(cmd, true)
	cmd = append(cmd,
		"-j", "DROP",
	)
	r.Holds = append(r.Holds, cmd)

	cmd = r.newCommand()
	if r.Interface != "host" {
		cmd = append(cmd,
			//"-m", "physdev",
			//"--physdev-out", r.Interface,
			"-o", r.Interface,
		)
	}
	cmd = r.commentCommand(cmd, true)
	cmd = append(cmd,
		"-j", "DROP",
	)
	r.Holds6 = append(r.Holds6, cmd)

	err = r.run(r.Holds, "-A", false)
	if err != nil {
		return
	}

	err = r.run(r.Holds6, "-A", true)
	if err != nil {
		return
	}

	return
}

func (r *Rules) Remove() (err error) {
	err = r.run(r.Ingress, "-D", false)
	if err != nil {
		return
	}
	r.Ingress = [][]string{}

	err = r.run(r.Ingress6, "-D", true)
	if err != nil {
		return
	}
	r.Ingress6 = [][]string{}

	err = r.run(r.Holds, "-D", false)
	if err != nil {
		return
	}
	r.Holds = [][]string{}

	err = r.run(r.Holds6, "-D", true)
	if err != nil {
		return
	}
	r.Holds6 = [][]string{}

	return
}

func generate(namespace, iface string, ingress []*firewall.Rule) (
	rules *Rules) {

	rules = &Rules{
		Namespace: namespace,
		Interface: iface,
		Ingress:   [][]string{},
		Ingress6:  [][]string{},
		Holds:     [][]string{},
		Holds6:    [][]string{},
	}

	cmd := rules.newCommand()
	if rules.Interface != "host" {
		//cmd = append(cmd,
		//	"-m", "physdev",
		//	"--physdev-out", rules.Interface,
		//)
		cmd = append(cmd,
			"-o", rules.Interface,
		)
	}
	cmd = append(cmd,
		"-m", "conntrack",
		"--ctstate", "RELATED,ESTABLISHED",
	)
	cmd = rules.commentCommand(cmd, false)
	cmd = append(cmd,
		"-j", "ACCEPT",
	)
	rules.Ingress = append(rules.Ingress, cmd)

	cmd = rules.newCommand()
	if rules.Interface != "host" {
		//cmd = append(cmd,
		//	"-m", "physdev",
		//	"--physdev-out", rules.Interface,
		//)
		cmd = append(cmd,
			"-o", rules.Interface,
		)
	}
	cmd = append(cmd,
		"-m", "conntrack",
		"--ctstate", "RELATED,ESTABLISHED",
	)
	cmd = rules.commentCommand(cmd, false)
	cmd = append(cmd,
		"-j", "ACCEPT",
	)
	rules.Ingress6 = append(rules.Ingress6, cmd)

	cmd = rules.newCommand()
	if rules.Interface != "host" {
		//cmd = append(cmd,
		//	"-m", "physdev",
		//	"--physdev-out", rules.Interface,
		//)
		cmd = append(cmd,
			"-o", rules.Interface,
		)
	}
	cmd = append(cmd,
		"-m", "conntrack",
		"--ctstate", "INVALID",
	)
	cmd = rules.commentCommand(cmd, false)
	cmd = append(cmd,
		"-j", "DROP",
	)
	rules.Ingress = append(rules.Ingress, cmd)

	cmd = rules.newCommand()
	if rules.Interface != "host" {
		//cmd = append(cmd,
		//	"-m", "physdev",
		//	"--physdev-out", rules.Interface,
		//)
		cmd = append(cmd,
			"-o", rules.Interface,
		)
	}
	cmd = append(cmd,
		"-m", "conntrack",
		"--ctstate", "INVALID",
	)
	cmd = rules.commentCommand(cmd, false)
	cmd = append(cmd,
		"-j", "DROP",
	)
	rules.Ingress6 = append(rules.Ingress6, cmd)

	if rules.Interface != "host" {
		cmd = rules.newCommand()
		cmd = append(cmd,
			//"-m", "physdev",
			//"--physdev-out", rules.Interface,
			"-o", rules.Interface,
			"-m", "pkttype",
			"--pkt-type", "multicast",
		)
		cmd = rules.commentCommand(cmd, false)
		cmd = append(cmd,
			"-j", "ACCEPT",
		)
		rules.Ingress = append(rules.Ingress, cmd)
	}

	if rules.Interface != "host" {
		cmd = rules.newCommand()
		cmd = append(cmd,
			//"-m", "physdev",
			//"--physdev-out", rules.Interface,
			"-o", rules.Interface,
			"-m", "pkttype",
			"--pkt-type", "broadcast",
		)
		cmd = rules.commentCommand(cmd, false)
		cmd = append(cmd,
			"-j", "ACCEPT",
		)
		rules.Ingress = append(rules.Ingress, cmd)
	}

	if rules.Interface != "host" {
		cmd = rules.newCommand()
		cmd = append(cmd,
			//"-m", "physdev",
			//"--physdev-out", rules.Interface,
			"-o", rules.Interface,
			"-m", "pkttype",
			"--pkt-type", "multicast",
		)
		cmd = rules.commentCommand(cmd, false)
		cmd = append(cmd,
			"-j", "ACCEPT",
		)
		rules.Ingress6 = append(rules.Ingress6, cmd)
	}

	if rules.Interface != "host" {
		cmd = rules.newCommand()
		cmd = append(cmd,
			//"-m", "physdev",
			//"--physdev-out", rules.Interface,
			"-o", rules.Interface,
			"-m", "pkttype",
			"--pkt-type", "broadcast",
		)
		cmd = rules.commentCommand(cmd, false)
		cmd = append(cmd,
			"-j", "ACCEPT",
		)
		rules.Ingress6 = append(rules.Ingress6, cmd)
	}

	if rules.Interface != "host" {
		cmd = rules.newCommand()
		cmd = append(cmd,
			"-o", rules.Interface,
			"-p", "ipv6-icmp",
			//"-m", "physdev",
			//"--physdev-out", rules.Interface,
		)
		cmd = rules.commentCommand(cmd, false)
		cmd = append(cmd,
			"-j", "ACCEPT",
		)
		rules.Ingress6 = append(rules.Ingress6, cmd)
	}

	for _, rule := range ingress {
		for _, sourceIp := range rule.SourceIps {
			ipv6 := strings.Contains(sourceIp, ":")
			cmd = rules.newCommand()

			if sourceIp != "0.0.0.0/0" && sourceIp != "::/0" {
				cmd = append(cmd,
					"-s", sourceIp,
				)
			}

			if rules.Interface != "host" {
				cmd = append(cmd,
					"-o", rules.Interface,
				)
			}

			switch rule.Protocol {
			case firewall.All:
				break
			case firewall.Icmp:
				if ipv6 {
					cmd = append(cmd,
						"-p", "ipv6-icmp",
					)
				} else {
					cmd = append(cmd,
						"-p", "icmp",
					)
				}
				break
			case firewall.Tcp, firewall.Udp:
				cmd = append(cmd,
					"-p", rule.Protocol,
				)
				break
			default:
				continue
			}

			if rules.Interface != "host" {
				//cmd = append(cmd,
				//	"-m", "physdev",
				//	"--physdev-out", rules.Interface,
				//)
			}

			switch rule.Protocol {
			case firewall.Tcp, firewall.Udp:
				cmd = append(cmd,
					"-m", rule.Protocol,
					"--dport", strings.Replace(rule.Port, "-", ":", 1),
				)
				break
			}

			cmd = rules.commentCommand(cmd, false)
			cmd = append(cmd,
				"-j", "ACCEPT",
			)

			if ipv6 {
				rules.Ingress6 = append(rules.Ingress6, cmd)
			} else {
				rules.Ingress = append(rules.Ingress, cmd)
			}
		}
	}

	cmd = rules.newCommand()
	if rules.Interface != "host" {
		//cmd = append(cmd,
		//	"-m", "physdev",
		//	"--physdev-out", rules.Interface,
		//)
		cmd = append(cmd,
			"-o", rules.Interface,
		)
	}
	cmd = rules.commentCommand(cmd, false)
	cmd = append(cmd,
		"-j", "DROP",
	)
	rules.Ingress = append(rules.Ingress, cmd)

	cmd = rules.newCommand()
	if rules.Interface != "host" {
		//cmd = append(cmd,
		//	"-m", "physdev",
		//	"--physdev-out", rules.Interface,
		//)
		cmd = append(cmd,
			"-o", rules.Interface,
		)
	}
	cmd = rules.commentCommand(cmd, false)
	cmd = append(cmd,
		"-j", "DROP",
	)
	rules.Ingress6 = append(rules.Ingress6, cmd)

	return
}

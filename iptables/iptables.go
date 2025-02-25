package iptables

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/firewall"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vpc"
	"github.com/sirupsen/logrus"
)

var (
	curState  *State
	stateLock = utils.NewTimeoutLock(3 * time.Minute)
)

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

func (r *Rules) newCommandNatPre() (cmd []string) {
	cmd = []string{
		"PREROUTING",
	}

	return
}

func (r *Rules) newCommandNatPost() (cmd []string) {
	cmd = []string{
		"POSTROUTING",
	}

	return
}

func (r *Rules) newCommandMap() (cmd []string) {
	cmd = []string{
		"PREROUTING",
	}

	return
}

func (r *Rules) newCommandMapPost() (cmd []string) {
	cmd = []string{
		"POSTROUTING",
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

func (r *Rules) commentCommandHeader(inCmd []string) (cmd []string) {
	cmd = append(inCmd,
		"-m", "comment",
		"--comment", "pritunl_cloud_head",
	)

	return
}

func (r *Rules) commentCommandSdc(inCmd []string) (cmd []string) {
	cmd = append(inCmd,
		"-m", "comment",
		"--comment", "pritunl_cloud_sdc",
	)

	return
}

func (r *Rules) commentCommandNat(inCmd []string) (cmd []string) {
	cmd = append(inCmd,
		"-m", "comment",
		"--comment", "pritunl_cloud_nat",
	)

	return
}

func (r *Rules) commentCommandMap(inCmd []string) (cmd []string) {
	cmd = append(inCmd,
		"-m", "comment",
		"--comment", "pritunl_cloud_map",
	)

	return
}

func (r *Rules) run(table string, cmds [][]string,
	ipCmd string, ipv6 bool) (err error) {

	iptablesCmd := getIptablesCmd(ipv6)

	for _, cmd := range cmds {
		cmd = append([]string{ipCmd}, cmd...)
		if table != "" {
			cmd = append([]string{"-t", table}, cmd...)
		}

		if r.Namespace != "0" {
			cmd = append([]string{
				"netns", "exec", r.Namespace,
				iptablesCmd,
			}, cmd...)
		}

		for i := 0; i < 3; i++ {
			output := ""
			if r.Namespace == "0" {
				Lock()
				output, err = utils.ExecCombinedOutputLogged(
					[]string{
						"matching rule exist",
					}, iptablesCmd, cmd...)
				Unlock()
			} else {
				Lock()
				output, err = utils.ExecCombinedOutputLogged(
					[]string{
						"matching rule exist",
						"Cannot open network namespace",
					}, "ip", cmd...)
				Unlock()
			}

			if err != nil {
				if i < 2 {
					err = nil
					time.Sleep(250 * time.Millisecond)
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

func (r *Rules) Apply(all bool) (err error) {
	if all || r.HeaderDiff {
		err = r.run("", r.Header, "-A", false)
		if err != nil {
			return
		}
	}

	if all || r.Header6Diff {
		err = r.run("", r.Header6, "-A", true)
		if err != nil {
			return
		}
	}

	if all || r.SourceDestCheckDiff {
		err = r.run("", r.SourceDestCheck, "-A", false)
		if err != nil {
			return
		}
	}

	if all || r.SourceDestCheck6Diff {
		err = r.run("", r.SourceDestCheck6, "-A", true)
		if err != nil {
			return
		}
	}

	if all || r.IngressDiff {
		err = r.run("", r.Ingress, "-A", false)
		if err != nil {
			return
		}
	}

	if all || r.Ingress6Diff {
		err = r.run("", r.Ingress6, "-A", true)
		if err != nil {
			return
		}
	}

	if all || r.NatsDiff {
		err = r.run("nat", r.Nats, "-A", false)
		if err != nil {
			return
		}
	}

	if all || r.Nats6Diff {
		err = r.run("nat", r.Nats6, "-A", true)
		if err != nil {
			return
		}
	}

	if all || r.MapsDiff {
		err = r.run("nat", r.Maps, "-A", false)
		if err != nil {
			return
		}
	}

	if all || r.Maps6Diff {
		err = r.run("nat", r.Maps6, "-A", true)
		if err != nil {
			return
		}
	}

	if all || r.IngressDiff {
		err = r.run("", r.Holds, "-D", false)
		if err != nil {
			return
		}
		r.Holds = [][]string{}
	}

	if all || r.Ingress6Diff {
		err = r.run("", r.Holds6, "-D", true)
		if err != nil {
			return
		}
		r.Holds6 = [][]string{}
	}

	return
}

func (r *Rules) Hold() (err error) {
	cmd := r.newCommand()
	if r.Interface != "host" {
		if strings.HasPrefix(r.Interface, "e") {
			cmd = append(cmd,
				"-i", r.Interface,
			)
		} else if strings.HasPrefix(r.Interface, "h") {
			cmd = append(cmd,
				"-i", r.Interface,
			)
		} else if strings.HasPrefix(r.Interface, "m") {
			cmd = append(cmd,
				"-i", r.Interface,
			)
		} else if strings.HasPrefix(r.Interface, "i") {
			cmd = append(cmd,
				"-i", r.Interface,
			)
		} else if strings.HasPrefix(r.Interface, "o") {
			cmd = append(cmd,
				"-i", r.Interface,
			)
		} else if strings.HasPrefix(r.Interface, "p") {
			cmd = append(cmd,
				"-m", "physdev",
				"--physdev-out", r.Interface,
				"--physdev-is-bridged",
			)
		} else {
			err = &errortypes.ParseError{
				errors.Newf("iptables: Unknown interface type %s",
					r.Interface),
			}
			return
		}
	}
	cmd = r.commentCommand(cmd, true)
	cmd = append(cmd,
		"-j", "DROP",
	)
	r.Holds = append(r.Holds, cmd)

	cmd = r.newCommand()
	if r.Interface != "host" {
		if strings.HasPrefix(r.Interface, "e") {
			cmd = append(cmd,
				"-i", r.Interface,
			)
		} else if strings.HasPrefix(r.Interface, "h") {
			cmd = append(cmd,
				"-i", r.Interface,
			)
		} else if strings.HasPrefix(r.Interface, "m") {
			cmd = append(cmd,
				"-i", r.Interface,
			)
		} else if strings.HasPrefix(r.Interface, "i") {
			cmd = append(cmd,
				"-i", r.Interface,
			)
		} else if strings.HasPrefix(r.Interface, "o") {
			cmd = append(cmd,
				"-i", r.Interface,
			)
		} else if strings.HasPrefix(r.Interface, "p") {
			cmd = append(cmd,
				"-m", "physdev",
				"--physdev-out", r.Interface,
				"--physdev-is-bridged",
			)
		} else {
			err = &errortypes.ParseError{
				errors.Newf("iptables: Unknown interface type %s",
					r.Interface),
			}
			return
		}
	}
	cmd = r.commentCommand(cmd, true)
	cmd = append(cmd,
		"-j", "DROP",
	)
	r.Holds6 = append(r.Holds6, cmd)

	err = r.run("", r.Holds, "-A", false)
	if err != nil {
		return
	}

	err = r.run("", r.Holds6, "-A", true)
	if err != nil {
		return
	}

	return
}

func (r *Rules) Remove(all bool) (err error) {
	if all || r.HeaderDiff {
		err = r.run("", r.Header, "-D", false)
		if err != nil {
			return
		}
		r.Header = [][]string{}
	}

	if all || r.Header6Diff {
		err = r.run("", r.Header6, "-D", true)
		if err != nil {
			return
		}
		r.Header6 = [][]string{}
	}

	if all || r.SourceDestCheckDiff {
		err = r.run("", r.SourceDestCheck, "-D", false)
		if err != nil {
			return
		}
		r.SourceDestCheck = [][]string{}
	}

	if all || r.SourceDestCheck6Diff {
		err = r.run("", r.SourceDestCheck6, "-D", true)
		if err != nil {
			return
		}
		r.SourceDestCheck6 = [][]string{}
	}

	if all || r.IngressDiff {
		err = r.run("", r.Ingress, "-D", false)
		if err != nil {
			return
		}
		r.Ingress = [][]string{}
	}

	if all || r.Ingress6Diff {
		err = r.run("", r.Ingress6, "-D", true)
		if err != nil {
			return
		}
		r.Ingress6 = [][]string{}
	}

	if all || r.NatsDiff {
		err = r.run("nat", r.Nats, "-D", false)
		if err != nil {
			return
		}
		r.Nats = [][]string{}
	}

	if all || r.Nats6Diff {
		err = r.run("nat", r.Nats6, "-D", true)
		if err != nil {
			return
		}
		r.Nats6 = [][]string{}
	}

	if all || r.MapsDiff {
		err = r.run("nat", r.Maps, "-D", false)
		if err != nil {
			return
		}
		r.Maps = [][]string{}
	}

	if all || r.Maps6Diff {
		err = r.run("nat", r.Maps6, "-D", true)
		if err != nil {
			return
		}
		r.Maps6 = [][]string{}
	}

	err = r.run("", r.Holds, "-D", false)
	if err != nil {
		return
	}
	r.Holds = [][]string{}

	err = r.run("", r.Holds6, "-D", true)
	if err != nil {
		return
	}
	r.Holds6 = [][]string{}

	return
}

func generateVirt(vc *vpc.Vpc, namespace, iface, addr, addr6 string,
	sourceDestCheck bool, ingress []*firewall.Rule) (rules *Rules) {

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

	cmd := rules.newCommand()
	cmd = append(cmd,
		"-p", "ipv6-icmp",
	)
	if rules.Interface != "host" {
		cmd = append(cmd,
			"-m", "physdev",
			"--physdev-out", rules.Interface,
			"--physdev-is-bridged",
		)
	}
	cmd = append(cmd,
		"-m", "icmp6",
		"--icmpv6-type", "133",
	)
	cmd = rules.commentCommandHeader(cmd)
	cmd = append(cmd,
		"-j", "ACCEPT",
	)
	rules.Header6 = append(rules.Header6, cmd)

	cmd = rules.newCommand()
	cmd = append(cmd,
		"-p", "ipv6-icmp",
	)
	if rules.Interface != "host" {
		cmd = append(cmd,
			"-m", "physdev",
			"--physdev-out", rules.Interface,
			"--physdev-is-bridged",
		)
	}
	cmd = append(cmd,
		"-m", "icmp6",
		"--icmpv6-type", "134",
	)
	cmd = rules.commentCommandHeader(cmd)
	cmd = append(cmd,
		"-j", "ACCEPT",
	)
	rules.Header6 = append(rules.Header6, cmd)

	cmd = rules.newCommand()
	cmd = append(cmd,
		"-p", "ipv6-icmp",
	)
	if rules.Interface != "host" {
		cmd = append(cmd,
			"-m", "physdev",
			"--physdev-out", rules.Interface,
			"--physdev-is-bridged",
		)
	}
	cmd = append(cmd,
		"-m", "icmp6",
		"--icmpv6-type", "135",
	)
	cmd = rules.commentCommandHeader(cmd)
	cmd = append(cmd,
		"-j", "ACCEPT",
	)
	rules.Header6 = append(rules.Header6, cmd)

	cmd = rules.newCommand()
	cmd = append(cmd,
		"-p", "ipv6-icmp",
	)
	if rules.Interface != "host" {
		cmd = append(cmd,
			"-m", "physdev",
			"--physdev-out", rules.Interface,
			"--physdev-is-bridged",
		)
	}
	cmd = append(cmd,
		"-m", "icmp6",
		"--icmpv6-type", "136",
	)
	cmd = rules.commentCommandHeader(cmd)
	cmd = append(cmd,
		"-j", "ACCEPT",
	)
	rules.Header6 = append(rules.Header6, cmd)

	if sourceDestCheck {
		if addr != "" {
			cmd := rules.newCommand()
			cmd = append(cmd,
				"!", "-s", addr+"/32",
			)
			if rules.Interface != "host" {
				cmd = append(cmd,
					"-m", "physdev",
					"--physdev-in", rules.Interface,
					"--physdev-is-bridged",
				)
			}
			cmd = rules.commentCommandSdc(cmd)
			cmd = append(cmd,
				"-j", "DROP",
			)
			rules.SourceDestCheck = append(rules.SourceDestCheck, cmd)
		}

		cmd := rules.newCommand()
		cmd = append(cmd,
			"-m", "set",
			"!", "--match-set", "pr6_sdc", "src",
		)
		if rules.Interface != "host" {
			cmd = append(cmd,
				"-m", "physdev",
				"--physdev-in", rules.Interface,
				"--physdev-is-bridged",
			)
		}
		cmd = rules.commentCommandSdc(cmd)
		cmd = append(cmd,
			"-j", "DROP",
		)
		rules.SourceDestCheck6 = append(rules.SourceDestCheck6, cmd)

		if addr != "" {
			cmd = rules.newCommand()
			cmd = append(cmd,
				"!", "-d", addr+"/32",
			)
			if rules.Interface != "host" {
				cmd = append(cmd,
					"-m", "physdev",
					"--physdev-out", rules.Interface,
					"--physdev-is-bridged",
				)
			}
			cmd = rules.commentCommandSdc(cmd)
			cmd = append(cmd,
				"-j", "DROP",
			)
			rules.SourceDestCheck = append(rules.SourceDestCheck, cmd)
		}

		cmd = rules.newCommand()
		cmd = append(cmd,
			"-m", "set",
			"!", "--match-set", "pr6_sdc", "dst",
		)
		if rules.Interface != "host" {
			cmd = append(cmd,
				"-m", "physdev",
				"--physdev-out", rules.Interface,
				"--physdev-is-bridged",
			)
		}
		cmd = rules.commentCommandSdc(cmd)
		cmd = append(cmd,
			"-j", "DROP",
		)
		rules.SourceDestCheck6 = append(rules.SourceDestCheck6, cmd)
	}

	cmd = rules.newCommand()
	if rules.Interface != "host" {
		cmd = append(cmd,
			"-m", "physdev",
			"--physdev-out", rules.Interface,
			"--physdev-is-bridged",
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
		cmd = append(cmd,
			"-m", "physdev",
			"--physdev-out", rules.Interface,
			"--physdev-is-bridged",
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

	for _, rule := range ingress {
		all4 := false
		all6 := false
		set4 := false
		set6 := false
		setName := rule.SetName(false)
		setName6 := rule.SetName(true)

		if setName == "" || setName6 == "" {
			continue
		}

		for _, sourceIp := range rule.SourceIps {
			ipv6 := strings.Contains(sourceIp, ":")

			if sourceIp == "0.0.0.0/0" {
				if all4 {
					continue
				}
				all4 = true
			} else if sourceIp == "::/0" {
				if all6 {
					continue
				}
				all6 = true
			} else {
				if ipv6 {
					if set6 {
						continue
					}
					set6 = true
				} else {
					if set4 {
						continue
					}
					set4 = true
				}
			}

			cmd = rules.newCommand()

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
			case firewall.Multicast, firewall.Broadcast:
				cmd = append(cmd,
					"-p", "udp",
					"-m", "pkttype",
					"--pkt-type", rule.Protocol,
				)
				break
			case firewall.Tcp, firewall.Udp:
				cmd = append(cmd,
					"-p", rule.Protocol,
				)
				break
			default:
				continue
			}

			if sourceIp != "0.0.0.0/0" && sourceIp != "::/0" &&
				rule.Protocol != firewall.Multicast &&
				rule.Protocol != firewall.Broadcast {

				if ipv6 {
					cmd = append(cmd,
						"-m", "set",
						"--match-set", setName6, "src",
					)
				} else {
					cmd = append(cmd,
						"-m", "set",
						"--match-set", setName, "src",
					)
				}
			}

			if rules.Interface != "host" {
				cmd = append(cmd,
					"-m", "physdev",
					"--physdev-out", rules.Interface,
					"--physdev-is-bridged",
				)
			}

			switch rule.Protocol {
			case firewall.Multicast, firewall.Broadcast:
				cmd = append(cmd,
					"-m", "udp",
					"--dport", strings.Replace(rule.Port, "-", ":", 1),
				)
				break
			case firewall.Tcp, firewall.Udp:
				cmd = append(cmd,
					"-m", rule.Protocol,
					"--dport", strings.Replace(rule.Port, "-", ":", 1),
					"-m", "conntrack",
					"--ctstate", "NEW",
				)
				break
			}

			if rule.Protocol == firewall.Multicast ||
				rule.Protocol == firewall.Broadcast {

				cmd = rules.commentCommandHeader(cmd)
				cmd = append(cmd,
					"-j", "ACCEPT",
				)

				if ipv6 {
					rules.Header6 = append(rules.Header6, cmd)
				} else {
					rules.Header = append(rules.Header, cmd)
				}
			} else {
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
	}

	cmd = rules.newCommand()
	if rules.Interface != "host" {
		cmd = append(cmd,
			"-m", "physdev",
			"--physdev-out", rules.Interface,
			"--physdev-is-bridged",
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
		cmd = append(cmd,
			"-m", "physdev",
			"--physdev-out", rules.Interface,
			"--physdev-is-bridged",
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

	cmd = rules.newCommand()
	if rules.Interface != "host" {
		cmd = append(cmd,
			"-m", "physdev",
			"--physdev-out", rules.Interface,
			"--physdev-is-bridged",
		)
	}
	cmd = rules.commentCommand(cmd, false)
	cmd = append(cmd,
		"-j", "DROP",
	)
	rules.Ingress = append(rules.Ingress, cmd)

	cmd = rules.newCommand()
	if rules.Interface != "host" {
		cmd = append(cmd,
			"-m", "physdev",
			"--physdev-out", rules.Interface,
			"--physdev-is-bridged",
		)
	}
	cmd = rules.commentCommand(cmd, false)
	cmd = append(cmd,
		"-j", "DROP",
	)
	rules.Ingress6 = append(rules.Ingress6, cmd)

	if vc != nil && vc.Maps != nil {
		for _, mp := range vc.Maps {
			if mp.Type != vpc.Destination {
				continue
			}

			if strings.Contains(mp.Target, ":") {
				if addr6 != "" {
					cmd = rules.newCommandMap()
					cmd = append(cmd,
						"-s", addr+"/128",
						"-d", mp.Destination,
					)
					cmd = rules.commentCommandMap(cmd)
					cmd = append(cmd,
						"-j", "DNAT",
						"--to-destination", mp.Target,
					)
					rules.Maps6 = append(rules.Maps6, cmd)
				}
			} else {
				if addr != "" {
					cmd = rules.newCommandMap()
					cmd = append(cmd,
						"-s", addr+"/32",
						"-d", mp.Destination,
					)
					cmd = rules.commentCommandMap(cmd)
					cmd = append(cmd,
						"-j", "DNAT",
						"--to-destination", mp.Target,
					)
					rules.Maps = append(rules.Maps, cmd)
				}
			}
		}
	}

	return
}

func generateInternal(namespace, iface string, nat, nat6, dhcp, dhcp6 bool,
	natAddr, natPubAddr, natAddr6, natPubAddr6 string,
	ingress []*firewall.Rule) (rules *Rules) {

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

	if nat && natAddr != "" && natPubAddr != "" {
		cmd := rules.newCommandNatPre()
		cmd = append(cmd,
			"-d", natPubAddr+"/32",
		)
		cmd = rules.commentCommandNat(cmd)
		cmd = append(cmd,
			"-j", "DNAT",
			"--to-destination", natAddr,
		)
		rules.Nats = append(rules.Nats, cmd)

		cmd = rules.newCommandNatPost()
		cmd = append(cmd,
			"-s", natAddr+"/32",
			"-d", natAddr+"/32",
		)
		cmd = rules.commentCommandNat(cmd)
		cmd = append(cmd,
			"-j", "SNAT",
			"--to-source", natPubAddr,
		)
		rules.Nats = append(rules.Nats, cmd)

		cmd = rules.newCommandNatPost()
		cmd = append(cmd,
			"-s", natAddr+"/32",
			"-o", iface,
		)
		cmd = rules.commentCommandNat(cmd)
		cmd = append(cmd,
			"-j", "MASQUERADE",
		)
		rules.Nats = append(rules.Nats, cmd)
	}

	if nat6 && natAddr6 != "" && natPubAddr6 != "" {
		cmd := rules.newCommandNatPre()
		cmd = append(cmd,
			"-d", natPubAddr6+"/128",
		)
		cmd = rules.commentCommandNat(cmd)
		cmd = append(cmd,
			"-j", "DNAT",
			"--to-destination", natAddr6,
		)
		rules.Nats6 = append(rules.Nats6, cmd)

		cmd = rules.newCommandNatPost()
		cmd = append(cmd,
			"-s", natAddr6+"/128",
			"-d", natAddr6+"/128",
		)
		cmd = rules.commentCommandNat(cmd)
		cmd = append(cmd,
			"-j", "SNAT",
			"--to-source", natPubAddr6,
		)
		rules.Nats6 = append(rules.Nats6, cmd)

		cmd = rules.newCommandNatPost()
		cmd = append(cmd,
			"-s", natAddr6+"/128",
			"-o", iface,
		)
		cmd = rules.commentCommandNat(cmd)
		cmd = append(cmd,
			"-j", "MASQUERADE",
		)
		rules.Nats6 = append(rules.Nats6, cmd)
	}

	cmd := rules.newCommand()
	if rules.Interface != "host" {
		cmd = append(cmd,
			"-i", rules.Interface,
		)
	}
	cmd = append(cmd,
		"-p", "ipv6-icmp",
		"-m", "icmp6",
		"--icmpv6-type", "133",
	)
	cmd = rules.commentCommandHeader(cmd)
	cmd = append(cmd,
		"-j", "ACCEPT",
	)
	rules.Header6 = append(rules.Header6, cmd)

	cmd = rules.newCommand()
	if rules.Interface != "host" {
		cmd = append(cmd,
			"-i", rules.Interface,
		)
	}
	cmd = append(cmd,
		"-p", "ipv6-icmp",
		"-m", "icmp6",
		"--icmpv6-type", "134",
	)
	cmd = rules.commentCommandHeader(cmd)
	cmd = append(cmd,
		"-j", "ACCEPT",
	)
	rules.Header6 = append(rules.Header6, cmd)

	cmd = rules.newCommand()
	if rules.Interface != "host" {
		cmd = append(cmd,
			"-i", rules.Interface,
		)
	}
	cmd = append(cmd,
		"-p", "ipv6-icmp",
		"-m", "icmp6",
		"--icmpv6-type", "135",
	)
	cmd = rules.commentCommandHeader(cmd)
	cmd = append(cmd,
		"-j", "ACCEPT",
	)
	rules.Header6 = append(rules.Header6, cmd)

	cmd = rules.newCommand()
	if rules.Interface != "host" {
		cmd = append(cmd,
			"-i", rules.Interface,
		)
	}
	cmd = append(cmd,
		"-p", "ipv6-icmp",
		"-m", "icmp6",
		"--icmpv6-type", "136",
	)
	cmd = rules.commentCommandHeader(cmd)
	cmd = append(cmd,
		"-j", "ACCEPT",
	)
	rules.Header6 = append(rules.Header6, cmd)

	cmd = rules.newCommand()
	if rules.Interface != "host" {
		cmd = append(cmd,
			"-i", rules.Interface,
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
		cmd = append(cmd,
			"-i", rules.Interface,
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

	for _, rule := range ingress {
		all4 := false
		all6 := false
		set4 := false
		set6 := false
		setName := rule.SetName(false)
		setName6 := rule.SetName(true)

		if setName == "" || setName6 == "" {
			continue
		}

		for _, sourceIp := range rule.SourceIps {
			ipv6 := strings.Contains(sourceIp, ":")

			if sourceIp == "0.0.0.0/0" {
				if all4 {
					continue
				}
				all4 = true
			} else if sourceIp == "::/0" {
				if all6 {
					continue
				}
				all6 = true
			} else {
				if ipv6 {
					if set6 {
						continue
					}
					set6 = true
				} else {
					if set4 {
						continue
					}
					set4 = true
				}
			}

			cmd = rules.newCommand()

			if rules.Interface != "host" {
				cmd = append(cmd,
					"-i", rules.Interface,
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
			case firewall.Multicast, firewall.Broadcast:
				cmd = append(cmd,
					"-p", "udp",
					"-m", "pkttype",
					"--pkt-type", rule.Protocol,
				)
				break
			case firewall.Tcp, firewall.Udp:
				cmd = append(cmd,
					"-p", rule.Protocol,
				)
				break
			default:
				continue
			}

			if sourceIp != "0.0.0.0/0" && sourceIp != "::/0" &&
				rule.Protocol != firewall.Multicast &&
				rule.Protocol != firewall.Broadcast {

				if ipv6 {
					cmd = append(cmd,
						"-m", "set",
						"--match-set", setName6, "src",
					)
				} else {
					cmd = append(cmd,
						"-m", "set",
						"--match-set", setName, "src",
					)
				}
			}

			switch rule.Protocol {
			case firewall.Multicast, firewall.Broadcast:
				cmd = append(cmd,
					"-m", "udp",
					"--dport", strings.Replace(rule.Port, "-", ":", 1),
				)
				break
			case firewall.Tcp, firewall.Udp:
				cmd = append(cmd,
					"-m", rule.Protocol,
					"--dport", strings.Replace(rule.Port, "-", ":", 1),
					"-m", "conntrack",
					"--ctstate", "NEW",
				)
				break
			}

			if rule.Protocol == firewall.Multicast ||
				rule.Protocol == firewall.Broadcast {

				cmd = rules.commentCommandHeader(cmd)
				cmd = append(cmd,
					"-j", "ACCEPT",
				)

				if ipv6 {
					rules.Header6 = append(rules.Header6, cmd)
				} else {
					rules.Header = append(rules.Header, cmd)
				}
			} else {
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
	}

	cmd = rules.newCommand()
	if rules.Interface != "host" {
		cmd = append(cmd,
			"-i", rules.Interface,
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
		cmd = append(cmd,
			"-i", rules.Interface,
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

	cmd = rules.newCommand()
	if rules.Interface != "host" {
		cmd = append(cmd,
			"-i", rules.Interface,
		)
	}
	cmd = rules.commentCommand(cmd, false)
	cmd = append(cmd,
		"-j", "DROP",
	)
	rules.Ingress = append(rules.Ingress, cmd)

	cmd = rules.newCommand()
	if rules.Interface != "host" {
		cmd = append(cmd,
			"-i", rules.Interface,
		)
	}
	cmd = rules.commentCommand(cmd, false)
	cmd = append(cmd,
		"-j", "DROP",
	)
	rules.Ingress6 = append(rules.Ingress6, cmd)

	return
}

func generateNodePort(namespace, iface string, addr, nodePortGateway string,
	firewallMaps []*firewall.Mapping) (rules *Rules) {

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

	cmd := rules.newCommand()
	if rules.Interface != "host" {
		cmd = append(cmd,
			"-i", rules.Interface,
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

	for _, mapping := range firewallMaps {
		if mapping.Protocol != firewall.Tcp &&
			mapping.Protocol != firewall.Udp {

			continue
		}

		cmd = rules.newCommandMap()
		cmd = append(cmd,
			"-i", iface,
			"-p", mapping.Protocol,
			"-m", mapping.Protocol,
			"--dport", fmt.Sprintf("%d", mapping.ExternalPort),
		)
		cmd = rules.commentCommandMap(cmd)
		cmd = append(cmd,
			"-j", "DNAT",
			"--to-destination", fmt.Sprintf(
				"%s:%d",
				addr,
				mapping.InternalPort,
			),
		)
		rules.Maps = append(rules.Maps, cmd)

		cmd = rules.newCommand()
		cmd = append(cmd,
			"-s", nodePortGateway+"/32",
		)

		if rules.Interface != "host" {
			cmd = append(cmd,
				"-i", rules.Interface,
			)
		}

		cmd = append(cmd,
			"-p", mapping.Protocol,
			"-m", mapping.Protocol,
			"--dport", fmt.Sprintf("%d", mapping.InternalPort),
			"-m", "conntrack",
			"--ctstate", "NEW",
		)

		cmd = rules.commentCommand(cmd, false)
		cmd = append(cmd,
			"-j", "ACCEPT",
		)

		rules.Ingress = append(rules.Ingress, cmd)
	}

	cmd = rules.newCommand()
	if rules.Interface != "host" {
		cmd = append(cmd,
			"-i", rules.Interface,
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
		cmd = append(cmd,
			"-i", rules.Interface,
		)
	}
	cmd = rules.commentCommand(cmd, false)
	cmd = append(cmd,
		"-j", "DROP",
	)
	rules.Ingress = append(rules.Ingress, cmd)

	return
}

func generateHost(namespace, iface string, nodePortNetwork bool,
	nodePortGateway string, externalIfaces []string, ingress []*firewall.Rule,
	nodePortMappings map[string][]*firewall.Mapping) (rules *Rules) {

	rules = &Rules{
		Namespace:        namespace,
		Interface:        iface,
		Header:           [][]string{},
		Header6:          [][]string{},
		SourceDestCheck:  [][]string{},
		SourceDestCheck6: [][]string{},
		Ingress:          [][]string{},
		Ingress6:         [][]string{},
		Holds:            [][]string{},
		Holds6:           [][]string{},
	}

	if rules.Interface == "host" {
		cmd := rules.newCommand()
		cmd = append(cmd,
			"-i", "lo",
		)
		cmd = rules.commentCommand(cmd, false)
		cmd = append(cmd,
			"-j", "ACCEPT",
		)
		rules.Ingress = append(rules.Ingress, cmd)
	}

	if rules.Interface == "host" {
		cmd := rules.newCommand()
		cmd = append(cmd,
			"-i", "lo",
		)
		cmd = rules.commentCommand(cmd, false)
		cmd = append(cmd,
			"-j", "ACCEPT",
		)
		rules.Ingress6 = append(rules.Ingress6, cmd)
	}

	cmd := rules.newCommand()
	if rules.Interface != "host" {
		cmd = append(cmd,
			"-o", rules.Interface,
		)
	}
	cmd = append(cmd,
		"-p", "ipv6-icmp",
		"-m", "icmp6",
		"--icmpv6-type", "133",
	)
	cmd = rules.commentCommandHeader(cmd)
	cmd = append(cmd,
		"-j", "ACCEPT",
	)
	rules.Header6 = append(rules.Header6, cmd)

	cmd = rules.newCommand()
	if rules.Interface != "host" {
		cmd = append(cmd,
			"-o", rules.Interface,
		)
	}
	cmd = append(cmd,
		"-p", "ipv6-icmp",
		"-m", "icmp6",
		"--icmpv6-type", "134",
	)
	cmd = rules.commentCommandHeader(cmd)
	cmd = append(cmd,
		"-j", "ACCEPT",
	)
	rules.Header6 = append(rules.Header6, cmd)

	cmd = rules.newCommand()
	if rules.Interface != "host" {
		cmd = append(cmd,
			"-o", rules.Interface,
		)
	}
	cmd = append(cmd,
		"-p", "ipv6-icmp",
		"-m", "icmp6",
		"--icmpv6-type", "135",
	)
	cmd = rules.commentCommandHeader(cmd)
	cmd = append(cmd,
		"-j", "ACCEPT",
	)
	rules.Header6 = append(rules.Header6, cmd)

	cmd = rules.newCommand()
	if rules.Interface != "host" {
		cmd = append(cmd,
			"-o", rules.Interface,
		)
	}
	cmd = append(cmd,
		"-p", "ipv6-icmp",
		"-m", "icmp6",
		"--icmpv6-type", "136",
	)
	cmd = rules.commentCommandHeader(cmd)
	cmd = append(cmd,
		"-j", "ACCEPT",
	)
	rules.Header6 = append(rules.Header6, cmd)

	cmd = rules.newCommand()
	if rules.Interface != "host" {
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

	for _, rule := range ingress {
		all4 := false
		all6 := false
		set4 := false
		set6 := false
		setName := rule.SetName(false)
		setName6 := rule.SetName(true)

		if setName == "" || setName6 == "" {
			continue
		}

		for _, sourceIp := range rule.SourceIps {
			ipv6 := strings.Contains(sourceIp, ":")

			if sourceIp == "0.0.0.0/0" {
				if all4 {
					continue
				}
				all4 = true
			} else if sourceIp == "::/0" {
				if all6 {
					continue
				}
				all6 = true
			} else {
				if ipv6 {
					if set6 {
						continue
					}
					set6 = true
				} else {
					if set4 {
						continue
					}
					set4 = true
				}
			}

			cmd = rules.newCommand()

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
			case firewall.Multicast, firewall.Broadcast:
				cmd = append(cmd,
					"-p", "udp",
					"-m", "pkttype",
					"--pkt-type", rule.Protocol,
				)
				break
			case firewall.Tcp, firewall.Udp:
				cmd = append(cmd,
					"-p", rule.Protocol,
				)
				break
			default:
				continue
			}

			if sourceIp != "0.0.0.0/0" && sourceIp != "::/0" &&
				rule.Protocol != firewall.Multicast &&
				rule.Protocol != firewall.Broadcast {

				if ipv6 {
					cmd = append(cmd,
						"-m", "set",
						"--match-set", setName6, "src",
					)
				} else {
					cmd = append(cmd,
						"-m", "set",
						"--match-set", setName, "src",
					)
				}
			}

			switch rule.Protocol {
			case firewall.Multicast, firewall.Broadcast:
				cmd = append(cmd,
					"-m", "udp",
					"--dport", strings.Replace(rule.Port, "-", ":", 1),
				)
				break
			case firewall.Tcp, firewall.Udp:
				cmd = append(cmd,
					"-m", rule.Protocol,
					"--dport", strings.Replace(rule.Port, "-", ":", 1),
					"-m", "conntrack",
					"--ctstate", "NEW",
				)
				break
			}

			if rule.Protocol == firewall.Multicast ||
				rule.Protocol == firewall.Broadcast {

				cmd = rules.commentCommandHeader(cmd)
				cmd = append(cmd,
					"-j", "ACCEPT",
				)

				if ipv6 {
					rules.Header6 = append(rules.Header6, cmd)
				} else {
					rules.Header = append(rules.Header, cmd)
				}
			} else {
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
	}

	cmd = rules.newCommand()
	if rules.Interface != "host" {
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

	cmd = rules.newCommand()
	if rules.Interface != "host" {
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
		cmd = append(cmd,
			"-o", rules.Interface,
		)
	}
	cmd = rules.commentCommand(cmd, false)
	cmd = append(cmd,
		"-j", "DROP",
	)
	rules.Ingress6 = append(rules.Ingress6, cmd)

	if nodePortNetwork {
		nodePortKeys := make([]string, 0, len(nodePortMappings))
		for k := range nodePortMappings {
			nodePortKeys = append(nodePortKeys, k)
		}
		sort.Strings(nodePortKeys)

		for _, nodePortAddr := range nodePortKeys {
			mappings := nodePortMappings[nodePortAddr]
			for _, mapping := range mappings {
				if mapping.Protocol != firewall.Tcp &&
					mapping.Protocol != firewall.Udp {
					continue
				}
				for _, externalIface := range externalIfaces {
					cmd = rules.newCommandMap()
					cmd = append(cmd,
						"-i", externalIface,
						"-p", mapping.Protocol,
						"-m", mapping.Protocol,
						"--dport", fmt.Sprintf("%d", mapping.ExternalPort),
					)
					cmd = rules.commentCommandMap(cmd)
					cmd = append(cmd,
						"-j", "DNAT",
						"--to-destination", fmt.Sprintf(
							"%s:%d",
							nodePortAddr,
							mapping.ExternalPort,
						),
					)
					rules.Maps = append(rules.Maps, cmd)
				}
			}
		}

		cmd = rules.newCommandMapPost()
		cmd = append(cmd,
			"-o", settings.Hypervisor.NodePortNetworkName,
		)
		cmd = rules.commentCommandMap(cmd)
		cmd = append(cmd,
			"-j", "SNAT",
			"--to-source", nodePortGateway,
		)
		rules.Maps = append(rules.Maps, cmd)
	}

	return
}

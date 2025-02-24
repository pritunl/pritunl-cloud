package hnetwork

import (
	"strings"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/block"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/state"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/sirupsen/logrus"
)

var (
	initialized = false
	curGateway  = ""
	curRule     *IptablesRule
)

type IptablesRule struct {
	Source string
	Output string
}

func (h *IptablesRule) Add() (err error) {
	args := []string{
		"-t", "nat",
		"-A", "POSTROUTING",
	}

	if h.Source != "" {
		args = append(args, "-s", h.Source)
	}
	if h.Output != "" {
		args = append(args, "-o", h.Output)
	}

	args = append(args,
		"-m", "comment",
		"--comment", "pritunl_cloud_host_nat",
		"-j", "MASQUERADE",
	)

	_, err = utils.ExecCombinedOutputLogged(
		[]string{
			"matching rule exist",
			"match by that name",
		},
		"iptables",
		args...,
	)
	if err != nil {
		return
	}
	if err != nil {
		return
	}

	return
}

func (h *IptablesRule) Remove() (err error) {
	args := []string{
		"-t", "nat",
		"-D", "POSTROUTING",
	}

	if h.Source != "" {
		args = append(args, "-s", h.Source)
	}
	if h.Output != "" {
		args = append(args, "-o", h.Output)
	}

	args = append(args,
		"-m", "comment",
		"--comment", "pritunl_cloud_host_nat",
		"-j", "MASQUERADE",
	)

	_, err = utils.ExecCombinedOutputLogged(
		[]string{
			"matching rule exist",
			"match by that name",
		},
		"iptables",
		args...,
	)
	if err != nil {
		return
	}

	return
}

func loadIptablesNat() (rules []*IptablesRule, err error) {
	rules = []*IptablesRule{}

	output, err := utils.ExecOutput("", "iptables", "-t", "nat", "-S")
	if err != nil {
		return
	}

	for _, line := range strings.Split(output, "\n") {
		if !strings.Contains(line, "POSTROUTING") ||
			!strings.Contains(line, "MASQUERADE") ||
			!strings.Contains(line, "pritunl_cloud_host_nat") {

			continue
		}

		cmd := strings.Fields(line)
		cmdLen := len(cmd)
		if cmdLen < 3 {
			logrus.WithFields(logrus.Fields{
				"iptables_rule": line,
			}).Error("hnetwork: Invalid iptables state")

			err = &errortypes.ParseError{
				errors.New("hnetwork: Invalid iptables state"),
			}
			return
		}

		rule := &IptablesRule{}

		for i, item := range cmd {
			if item == "-s" {
				if len(cmd) < i+2 {
					logrus.WithFields(logrus.Fields{
						"iptables_rule": line,
					}).Error("hnetwork: Invalid iptables host nat source")

					err = &errortypes.ParseError{
						errors.New(
							"hnetwork: Invalid iptables host nat source"),
					}
					return
				}
				rule.Source = cmd[i+1]
			}

			if item == "-o" {
				if len(cmd) < i+2 {
					logrus.WithFields(logrus.Fields{
						"iptables_rule": line,
					}).Error("hnetwork: Invalid iptables host nat output")

					err = &errortypes.ParseError{
						errors.New(
							"hnetwork: Invalid iptables host nat output"),
					}
					return
				}
				rule.Output = cmd[i+1]
			}
		}

		rules = append(rules, rule)
	}

	return
}

func removeNetwork(stat *state.State) (err error) {
	if curGateway != "" || stat.HasInterfaces(
		settings.Hypervisor.HostNetworkName) {

		err = clearAddr()
		if err != nil {
			return
		}

		curGateway = ""
	}

	return
}

func ApplyState(stat *state.State) (err error) {
	if !initialized {
		addr, e := getAddr()
		if e != nil {
			err = e
			return
		}

		rules, e := loadIptablesNat()
		if e != nil {
			err = e
			return
		}

		if len(rules) > 1 {
			for _, rule := range rules {
				err = rule.Remove()
				if err != nil {
					return
				}
			}
		} else if len(rules) == 1 {
			curRule = rules[0]
		}

		initialized = true
		curGateway = addr
	}

	if !stat.HasInterfaces(settings.Hypervisor.HostNetworkName) {
		logrus.WithFields(logrus.Fields{
			"iface": settings.Hypervisor.HostNetworkName,
		}).Info("hnetwork: Creating host interface")

		err = create()
		if err != nil {
			return
		}
	}

	hostBlock, err := block.GetNodeBlock(stat.Node().Id)
	if err != nil {
		return
	}

	gatewayCidr := hostBlock.GetGatewayCidr()
	if gatewayCidr == "" {
		logrus.WithFields(logrus.Fields{
			"host_block": hostBlock.Id.Hex(),
		}).Error("hnetwork: Host network block gateway is invalid")

		err = removeNetwork(stat)
		if err != nil {
			return
		}

		return
	}

	if curGateway != gatewayCidr {
		logrus.WithFields(logrus.Fields{
			"host_block":         hostBlock.Id.Hex(),
			"host_block_gateway": gatewayCidr,
		}).Info("hnetwork: Updating host network bridge")

		err = setAddr(gatewayCidr)
		if err != nil {
			return
		}

		curGateway = gatewayCidr
	}

	if stat.Node().HostNat {
		hostNet, e := hostBlock.GetNetwork()
		if e != nil {
			logrus.WithFields(logrus.Fields{
				"host_block": hostBlock.Id.Hex(),
				"error":      e,
			}).Error("hnetwork: Host nat block network invalid")
		} else {
			newRule := &IptablesRule{
				Source: hostNet.String(),
				Output: stat.Node().DefaultInterface,
			}

			if curRule == nil || curRule.Source != newRule.Source ||
				curRule.Output != newRule.Output {

				logrus.WithFields(logrus.Fields{
					"host_block":  hostBlock.Id.Hex(),
					"host_source": newRule.Source,
					"host_output": newRule.Output,
				}).Info("hnetwork: Updating host network nat")

				if curRule != nil {
					err = curRule.Remove()
					if err != nil {
						return
					}
					curRule = nil
				}

				err = newRule.Add()
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"host_block":  hostBlock.Id.Hex(),
						"host_source": newRule.Source,
						"host_output": newRule.Output,
						"error":       err,
					}).Error("hnetwork: Host nat add rule failed")
					err = nil
				} else {
					curRule = newRule
				}
			}
		}
	} else if curRule != nil {
		logrus.WithFields(logrus.Fields{
			"host_block":  hostBlock.Id.Hex(),
			"host_source": curRule.Source,
			"host_output": curRule.Output,
		}).Info("hnetwork: Updating host network nat")

		err = curRule.Remove()
		if err != nil {
			return
		}
		curRule = nil
	}

	return
}

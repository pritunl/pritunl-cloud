package iptables

import (
	"github.com/pritunl/pritunl-cloud/firewall"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/sirupsen/logrus"
)

type State struct {
	Interfaces map[string]*Rules
}

func LoadState(nodeSelf *node.Node, instances []*instance.Instance,
	nodeFirewall []*firewall.Rule, firewalls map[string][]*firewall.Rule) (
	state *State) {

	nodeNetworkMode := node.Self.NetworkMode
	if nodeNetworkMode == "" {
		nodeNetworkMode = node.Dhcp
	}
	nodeNetworkMode6 := node.Self.NetworkMode6
	if nodeNetworkMode6 == "" {
		nodeNetworkMode6 = node.Dhcp
	}

	state = &State{
		Interfaces: map[string]*Rules{},
	}

	if nodeFirewall != nil {
		state.Interfaces["0-host"] = generate("0", "host", nodeFirewall)
	}

	hostNetwork := !nodeSelf.HostBlock.IsZero()

	for _, inst := range instances {
		if !inst.IsActive() {
			continue
		}

		namespace := vm.GetNamespace(inst.Id, 0)
		iface := vm.GetIface(inst.Id, 0)
		ifaceHost := vm.GetIfaceHost(inst.Id, 0)

		ifaceExternal := vm.GetIfaceExternal(inst.Id, 0)
		ifaceExternal6 := ifaceExternal

		if nodeNetworkMode == node.Disabled &&
			nodeNetworkMode6 != node.Disabled &&
			nodeNetworkMode6 != node.Oracle {

			ifaceExternal6 = vm.GetIfaceExternal(inst.Id, 1)
		}

		addr := ""
		addr6 := ""
		pubAddr := ""
		pubAddr6 := ""
		if inst.PrivateIps != nil && len(inst.PrivateIps) != 0 {
			addr = inst.PrivateIps[0]
		}
		if inst.PrivateIps6 != nil && len(inst.PrivateIps6) != 0 {
			addr6 = inst.PrivateIps6[0]
		}
		if inst.PublicIps != nil && len(inst.PublicIps) != 0 {
			pubAddr = inst.PublicIps[0]
		}
		if inst.PublicIps6 != nil && len(inst.PublicIps6) != 0 {
			pubAddr6 = inst.PublicIps6[0]
		}

		oracleAddr := ""
		oracleIface := vm.GetIfaceOracle(inst.Id, 0)
		if inst.OraclePrivateIps != nil && len(inst.OraclePrivateIps) != 0 {
			oracleAddr = inst.OraclePrivateIps[0]
		}

		_, ok := state.Interfaces[namespace+"-"+iface]
		if ok {
			logrus.WithFields(logrus.Fields{
				"namespace": namespace,
				"interface": iface,
			}).Error("iptables: Virtual interface conflict")
			continue
		}

		ingress := firewalls[namespace]
		if ingress == nil {
			logrus.WithFields(logrus.Fields{
				"instance_id": inst.Id.Hex(),
				"namespace":   namespace,
			}).Warn("iptables: Failed to load instance firewall rules")
			continue
		}

		// TODO Move to netconf

		nat6 := false
		if nodeNetworkMode6 != node.Disabled &&
			nodeNetworkMode6 != node.Oracle {

			if ifaceExternal != ifaceExternal6 {
				rules := generateInternal(namespace, ifaceExternal6,
					false, true, addr, pubAddr, addr6, pubAddr6,
					oracleAddr, ingress)
				state.Interfaces[namespace+"-"+ifaceExternal6] = rules
			} else {
				nat6 = true
			}
		}

		if nodeNetworkMode != node.Disabled &&
			nodeNetworkMode != node.Oracle {

			rules := generateInternal(namespace, ifaceExternal,
				true, nat6, addr, pubAddr, addr6, pubAddr6,
				oracleAddr, ingress)
			state.Interfaces[namespace+"-"+ifaceExternal] = rules
		}

		if nodeNetworkMode == node.Oracle {
			rules := generateInternal(namespace, oracleIface,
				true, false, addr, pubAddr, addr6, pubAddr6,
				oracleAddr, ingress)

			state.Interfaces[namespace+"-"+oracleIface] = rules
		}

		if hostNetwork {
			rules := generateInternal(namespace, ifaceHost,
				false, false, "", "", "", "", "", ingress)
			state.Interfaces[namespace+"-"+ifaceHost] = rules
		}

		rules := generateVirt(namespace, iface, addr, addr6,
			!inst.SkipSourceDestCheck, ingress)
		state.Interfaces[namespace+"-"+iface] = rules
	}

	return
}

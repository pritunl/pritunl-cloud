package iptables

import (
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/block"
	"github.com/pritunl/pritunl-cloud/firewall"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/pritunl/pritunl-cloud/vpc"
	"github.com/sirupsen/logrus"
)

type State struct {
	Interfaces map[string]*Rules
}

func LoadState(nodeSelf *node.Node, vpcs []*vpc.Vpc,
	instances []*instance.Instance, nodeFirewall []*firewall.Rule,
	firewalls map[string][]*firewall.Rule,
	firewallMaps map[string][]*firewall.Mapping) (state *State) {

	vpcsMap := map[primitive.ObjectID]*vpc.Vpc{}
	for _, vc := range vpcs {
		vpcsMap[vc.Id] = vc
	}

	nodeNetworkMode := node.Self.NetworkMode
	if nodeNetworkMode == "" {
		nodeNetworkMode = node.Dhcp
	}
	nodeNetworkMode6 := node.Self.NetworkMode6
	if nodeNetworkMode6 == "" {
		nodeNetworkMode6 = node.Dhcp
	}
	nodePortNetwork := !node.Self.NoNodePortNetwork

	state = &State{
		Interfaces: map[string]*Rules{},
	}

	hostNodePortMappings := map[string][]*firewall.Mapping{}
	nodePortGateway, err := block.GetNodePortGateway()
	if err != nil {
		return
	}

	for _, inst := range instances {
		if !inst.IsActive() {
			continue
		}

		namespace := vm.GetNamespace(inst.Id, 0)
		iface := vm.GetIface(inst.Id, 0)
		ifaceHost := vm.GetIfaceHost(inst.Id, 1)
		ifaceNodePort := vm.GetIfaceNodePort(inst.Id, 1)
		ifaceExternal := vm.GetIfaceExternal(inst.Id, 0)

		addr := ""
		addr6 := ""
		pubAddr := ""
		pubAddr6 := ""
		if len(inst.PrivateIps) != 0 {
			addr = inst.PrivateIps[0]
		}
		if len(inst.PrivateIps6) != 0 {
			addr6 = inst.PrivateIps6[0]
		}
		if len(inst.PublicIps) != 0 {
			pubAddr = inst.PublicIps[0]
		}
		if len(inst.PublicIps6) != 0 {
			pubAddr6 = inst.PublicIps6[0]
		} else if len(inst.OraclePublicIps6) != 0 {
			pubAddr6 = inst.OraclePublicIps6[0]
		}

		nodePortAddr := ""
		if len(inst.NodePortIps) != 0 {
			nodePortAddr = inst.NodePortIps[0]
		}
		hostNodePortMappings[nodePortAddr] = firewallMaps[namespace]

		oracleAddr := ""
		oracleIface := vm.GetIfaceOracle(inst.Id, 0)
		if len(inst.OraclePrivateIps) != 0 {
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

		dhcp := false
		dhcp6 := false

		if nodeNetworkMode == node.Dhcp {
			dhcp = true
		}
		if nodeNetworkMode6 == node.Dhcp {
			dhcp6 = true
		}

		nat6 := false
		if nodeNetworkMode != node.Disabled &&
			nodeNetworkMode != node.Oracle {

			if nodeNetworkMode6 != node.Disabled &&
				nodeNetworkMode6 != node.Oracle {

				nat6 = true
			}

			rules := generateInternal(namespace, ifaceExternal,
				true, nat6, dhcp, dhcp6, addr, pubAddr, addr6, pubAddr6,
				ingress)
			state.Interfaces[namespace+"-"+ifaceExternal] = rules
		} else if nodeNetworkMode6 != node.Disabled &&
			nodeNetworkMode6 != node.Oracle {

			rules := generateInternal(namespace, ifaceExternal,
				false, true, dhcp, dhcp6, addr, pubAddr, addr6, pubAddr6,
				ingress)
			state.Interfaces[namespace+"-"+ifaceExternal] = rules
		}

		if nodeNetworkMode == node.Oracle {
			if nodeNetworkMode6 == node.Oracle {
				nat6 = true
			}

			rules := generateInternal(namespace, oracleIface,
				true, nat6, false, false, addr, oracleAddr, addr6, pubAddr6,
				ingress)

			state.Interfaces[namespace+"-"+oracleIface] = rules
		}

		rules := generateInternal(namespace, ifaceHost,
			false, false, false, false, "", "", "", "", ingress)
		state.Interfaces[namespace+"-"+ifaceHost] = rules

		if nodePortNetwork {
			rules := generateNodePort(namespace, ifaceNodePort,
				addr, nodePortGateway, firewallMaps[namespace])
			state.Interfaces[namespace+"-"+ifaceNodePort] = rules
		}

		rules = generateVirt(vpcsMap[inst.Vpc], namespace, iface, addr,
			addr6, !inst.SkipSourceDestCheck, ingress)
		state.Interfaces[namespace+"-"+iface] = rules
	}

	if nodeFirewall != nil {
		state.Interfaces["0-host"] = generateHost("0", "host",
			!nodeSelf.NoNodePortNetwork, nodePortGateway,
			nodeSelf.ExternalInterfaces, nodeSelf.PublicIps,
			nodeFirewall, hostNodePortMappings)
	}

	return
}

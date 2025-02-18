package netconf

import (
	"fmt"
	"strconv"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/block"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/interfaces"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/pritunl/pritunl-cloud/zone"
)

func (n *NetConf) Iface1(db *database.Database) (err error) {
	zne, err := zone.Get(db, node.Self.Zone)
	if err != nil {
		return
	}

	n.Vxlan = false
	if zne.NetworkMode == zone.VxlanVlan {
		n.Vxlan = true
	}

	n.NetworkMode = node.Self.NetworkMode
	if n.NetworkMode == "" {
		n.NetworkMode = node.Dhcp
	}
	n.NetworkMode6 = node.Self.NetworkMode6
	if n.NetworkMode6 == "" {
		n.NetworkMode6 = node.Dhcp
	}
	if n.NetworkMode == node.Internal || n.Virt.NoPublicAddress {
		n.NetworkMode = node.Disabled
	}
	if n.Virt.NoPublicAddress6 {
		n.NetworkMode6 = node.Disabled
	}

	if !node.Self.NoHostNetwork && !n.Virt.NoHostAddress {
		n.HostNetwork = true
		if node.Self.HostNat {
			n.HostNat = true
		}

		blck, e := block.GetNodeBlock(node.Self.Id)
		if e != nil {
			err = e
			return
		}

		hostNetwork, e := blck.GetNetwork()
		if e != nil {
			err = e
			return
		}
		n.HostSubnet = hostNetwork.String()
	}

	if !node.Self.NoNodePortNetwork && len(n.Virt.NodePorts) > 0 {
		n.NodePortNetwork = true

		blck, e := block.GetNodePortBlock(node.Self.Id)
		if e != nil {
			err = e
			return
		}

		nodePortNetwork, e := blck.GetNetwork()
		if e != nil {
			err = e
			return
		}
		n.NodePortSubnet = nodePortNetwork.String()
	}

	n.OracleSubnets = set.NewSet()
	if node.Self.OracleSubnets != nil {
		for _, subnet := range node.Self.OracleSubnets {
			n.OracleSubnets.Add(subnet)
		}
	}

	n.JumboFramesExternal = node.Self.JumboFrames
	n.JumboFramesInternal = node.Self.JumboFrames ||
		node.Self.JumboFramesInternal

	n.Namespace = vm.GetNamespace(n.Virt.Id, 0)

	if n.Virt.NetworkAdapters == nil || len(n.Virt.NetworkAdapters) < 1 {
		err = &errortypes.ParseError{
			errors.New("netconf: Missing virt network adapter"),
		}
		return
	}
	n.VmAdapter = n.Virt.NetworkAdapters[0]

	n.SystemExternalIface = vm.GetIfaceNodeExternal(n.Virt.Id, 0)

	return
}

func (n *NetConf) Iface2(db *database.Database, clean bool) (err error) {
	n.PhysicalHostIface = settings.Hypervisor.HostNetworkName
	n.PhysicalNodePortIface = settings.Hypervisor.NodePortNetworkName

	n.VirtIface = vm.GetIface(n.Virt.Id, 0)
	n.SystemInternalIface = vm.GetIfaceNodeInternal(n.Virt.Id, 0)
	n.SystemHostIface = vm.GetIfaceHost(n.Virt.Id, 0)
	n.SystemNodePortIface = vm.GetIfaceNodePort(n.Virt.Id, 0)
	n.SpaceExternalIface = vm.GetIfaceExternal(n.Virt.Id, 0)
	n.SpaceInternalIface = vm.GetIfaceInternal(n.Virt.Id, 0)
	n.SpaceHostIface = vm.GetIfaceHost(n.Virt.Id, 1)
	n.SpaceNodePortIface = vm.GetIfaceNodePort(n.Virt.Id, 1)
	n.SpaceOracleIface = vm.GetIfaceOracle(n.Virt.Id, 0)
	n.SpaceOracleVirtIface = vm.GetIfaceOracleVirt(n.Virt.Id, 0)
	n.SpaceImdsIface = "imds0"

	n.BridgeInternalIface = vm.GetIfaceVlan(n.Virt.Id, 0)

	n.PhysicalInternalIface = interfaces.GetInternal(
		n.SystemInternalIface, n.Vxlan)

	n.DhcpPidPath = fmt.Sprintf(
		"/var/run/dhclient-%s.pid",
		n.SpaceExternalIface,
	)
	n.DhcpLeasePath = paths.GetLeasePath(n.Virt.Id)

	if n.JumboFramesExternal || n.JumboFramesInternal || n.Vxlan {
		mtuSizeExternal := 0
		mtuSizeInternal := 0

		if n.JumboFramesExternal {
			mtuSizeExternal = settings.Hypervisor.JumboMtu
		} else {
			mtuSizeExternal = settings.Hypervisor.NormalMtu
		}
		if n.JumboFramesInternal {
			mtuSizeInternal = settings.Hypervisor.JumboMtu
		} else {
			mtuSizeInternal = settings.Hypervisor.NormalMtu
		}

		n.SpaceExternalIfaceMtu = strconv.Itoa(mtuSizeExternal)
		n.SystemExternalIfaceMtu = strconv.Itoa(mtuSizeExternal)

		n.SpaceHostIfaceMtu = strconv.Itoa(mtuSizeInternal)
		n.SpaceNodePortIfaceMtu = strconv.Itoa(mtuSizeInternal)
		n.SystemHostIfaceMtu = strconv.Itoa(mtuSizeInternal)
		n.SystemNodePortIfaceMtu = strconv.Itoa(mtuSizeInternal)
		n.ImdsIfaceMtu = strconv.Itoa(mtuSizeInternal)

		if n.Vxlan {
			mtuSizeExternal -= 50
			mtuSizeInternal -= 50
		}

		n.SpaceInternalIfaceMtu = strconv.Itoa(mtuSizeInternal)
		n.BridgeInternalIfaceMtu = strconv.Itoa(mtuSizeInternal)
		n.SystemInternalIfaceMtu = strconv.Itoa(mtuSizeInternal)

		if n.Vxlan {
			mtuSizeExternal -= 4
			mtuSizeInternal -= 4
		}

		n.VirtIfaceMtu = strconv.Itoa(mtuSizeInternal)
	}

	return
}

package netconf

import (
	"strconv"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/block"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/datacenter"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/interfaces"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/vm"
)

func (n *NetConf) Iface1(db *database.Database) (err error) {
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

	if !node.Self.NoNodePortNetwork {
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

	n.CloudSubnets = set.NewSet()
	if node.Self.CloudSubnets != nil {
		for _, subnet := range node.Self.CloudSubnets {
			n.CloudSubnets.Add(subnet)
		}
	}

	n.Namespace = vm.GetNamespace(n.Virt.Id, 0)

	if n.Virt.NetworkAdapters == nil || len(n.Virt.NetworkAdapters) < 1 {
		err = &errortypes.ParseError{
			errors.New("netconf: Missing virt network adapter"),
		}
		return
	}
	n.VmAdapter = n.Virt.NetworkAdapters[0]

	n.VirtIface = vm.GetIface(n.Virt.Id, 0)
	n.SystemExternalIface = vm.GetIfaceNodeExternal(n.Virt.Id, 0)
	n.SystemInternalIface = vm.GetIfaceNodeInternal(n.Virt.Id, 0)
	n.SystemHostIface = vm.GetIfaceHost(n.Virt.Id, 0)
	n.SystemNodePortIface = vm.GetIfaceNodePort(n.Virt.Id, 0)
	n.SpaceExternalIface = vm.GetIfaceExternal(n.Virt.Id, 0)
	n.SpaceInternalIface = vm.GetIfaceInternal(n.Virt.Id, 0)
	n.SpaceHostIface = vm.GetIfaceHost(n.Virt.Id, 1)
	n.SpaceNodePortIface = vm.GetIfaceNodePort(n.Virt.Id, 1)
	n.SpaceCloudIface = vm.GetIfaceCloud(n.Virt.Id, 0)
	n.SpaceCloudVirtIface = vm.GetIfaceCloudVirt(n.Virt.Id, 0)
	n.SpaceBridgeIface = settings.Hypervisor.BridgeIfaceName
	n.SpaceImdsIface = settings.Hypervisor.ImdsIfaceName

	return
}

func (n *NetConf) Iface2(db *database.Database, clean bool) (err error) {
	dc, err := datacenter.Get(db, node.Self.Datacenter)
	if err != nil {
		return
	}

	n.Vxlan = dc.Vxlan()

	n.PhysicalHostIface = settings.Hypervisor.HostNetworkName
	n.PhysicalNodePortIface = settings.Hypervisor.NodePortNetworkName

	n.BridgeInternalIface = vm.GetIfaceVlan(n.Virt.Id, 0)

	n.PhysicalInternalIface = interfaces.GetInternal(
		n.SystemInternalIface, n.Vxlan)

	mtuSizeExternal := dc.GetBaseExternalMtu()
	mtuSizeInternal := dc.GetBaseInternalMtu()
	mtuSizeOverlay := dc.GetOverlayMtu()
	mtuSizeInstance := dc.GetInstanceMtu()

	n.SpaceExternalIfaceMtu = strconv.Itoa(mtuSizeExternal)
	n.SystemExternalIfaceMtu = strconv.Itoa(mtuSizeExternal)

	n.SpaceHostIfaceMtu = strconv.Itoa(mtuSizeInternal)
	n.SpaceNodePortIfaceMtu = strconv.Itoa(mtuSizeInternal)
	n.SystemHostIfaceMtu = strconv.Itoa(mtuSizeInternal)
	n.SystemNodePortIfaceMtu = strconv.Itoa(mtuSizeInternal)
	n.ImdsIfaceMtu = strconv.Itoa(mtuSizeInternal)

	n.SpaceInternalIfaceMtu = strconv.Itoa(mtuSizeOverlay)
	n.BridgeInternalIfaceMtu = strconv.Itoa(mtuSizeOverlay)
	n.SystemInternalIfaceMtu = strconv.Itoa(mtuSizeOverlay)

	n.VirtIfaceMtu = strconv.Itoa(mtuSizeInstance)

	return
}

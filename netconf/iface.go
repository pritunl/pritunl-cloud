package netconf

import (
	"fmt"
	"strconv"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/interfaces"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/pritunl/pritunl-cloud/zone"
)

func (n *NetConf) Iface(db *database.Database) (err error) {
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
		n.NetworkMode6 = node.Disabled
	}
	n.HostBlock = node.Self.HostBlock
	if !n.HostBlock.IsZero() && !n.Virt.NoHostAddress {
		n.HostNetwork = true
	}

	n.OracleSubnets = set.NewSet()
	if node.Self.OracleSubnets != nil {
		for _, subnet := range node.Self.OracleSubnets {
			n.OracleSubnets.Add(subnet)
		}
	}

	n.JumboFrames = node.Self.JumboFrames
	n.Namespace = vm.GetNamespace(n.Virt.Id, 0)
	n.VmAdapter = n.Virt.NetworkAdapters[0]
	n.PhysicalHostIface = settings.Hypervisor.HostNetworkName

	n.VirtIface = vm.GetIface(n.Virt.Id, 0)
	n.SystemExternalIface = vm.GetIfaceVirt(n.Virt.Id, 0)
	n.SystemInternalIface = vm.GetIfaceVirt(n.Virt.Id, 1)
	n.SystemHostIface = vm.GetIfaceVirt(n.Virt.Id, 2)
	n.SpaceExternalIface = vm.GetIfaceExternal(n.Virt.Id, 0)
	n.SpaceInternalIface = vm.GetIfaceInternal(n.Virt.Id, 0)
	n.SpaceHostIface = vm.GetIfaceHost(n.Virt.Id, 0)
	n.SpaceOracleIface = vm.GetIfaceOracle(n.Virt.Id, 0)
	n.SpaceOracleVirtIface = vm.GetIfaceOracleVirt(n.Virt.Id, 0)

	n.BridgeInternalIface = vm.GetIfaceVlan(n.Virt.Id, 0)

	n.PhysicalInternalIface = interfaces.GetInternal(
		n.SystemInternalIface, n.Vxlan)

	n.DhcpPidPath = fmt.Sprintf(
		"/var/run/dhclient-%s.pid",
		n.SpaceExternalIface,
	)
	n.DhcpLeasePath = paths.GetLeasePath(n.Virt.Id)

	n.SpaceExternalIface = n.SpaceExternalIface
	n.SystemExternalIface6 = n.SystemExternalIface
	if n.NetworkMode != n.NetworkMode6 ||
		n.NetworkMode6 == node.Static {

		n.SpaceExternalIface6 = vm.GetIfaceExternal(n.Virt.Id, 1)
		n.SystemExternalIface6 = vm.GetIfaceVirt(n.Virt.Id, 3)
	} else if n.NetworkMode == n.NetworkMode6 {
		n.SpaceExternalIface6 = n.SpaceExternalIface
		n.SystemExternalIface6 = n.SystemExternalIface
	}

	if n.JumboFrames || n.Vxlan {
		mtuSize := 0
		if n.JumboFrames {
			mtuSize = settings.Hypervisor.JumboMtu
		} else {
			mtuSize = settings.Hypervisor.NormalMtu
		}

		n.SpaceExternalIfaceMtu = strconv.Itoa(mtuSize)
		n.SpaceExternalIfaceMtu6 = strconv.Itoa(mtuSize)
		n.SystemExternalIfaceMtu = strconv.Itoa(mtuSize)
		n.SystemExternalIfaceMtu6 = strconv.Itoa(mtuSize)

		n.SpaceHostIfaceMtu = strconv.Itoa(mtuSize)
		n.SystemHostIfaceMtu = strconv.Itoa(mtuSize)

		if n.Vxlan {
			mtuSize -= 50
		}

		n.SpaceInternalIfaceMtu = strconv.Itoa(mtuSize)
		n.BridgeInternalIfaceMtu = strconv.Itoa(mtuSize)
		n.SystemInternalIfaceMtu = strconv.Itoa(mtuSize)

		if n.Vxlan {
			mtuSize -= 4
		}

		n.VirtIfaceMtu = strconv.Itoa(mtuSize)
	}

	return
}

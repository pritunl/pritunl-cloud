package netconf

import (
	"fmt"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/interfaces"
	"github.com/pritunl/pritunl-cloud/node"

	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/pritunl/pritunl-cloud/vpc"
)

func (n *NetConf) Address(db *database.Database) (err error) {
	vc, err := vpc.Get(db, n.VmAdapter.Vpc)
	if err != nil {
		return
	}

	n.VxlanId = vc.VpcId

	vcNet, err := vc.GetNetwork()
	if err != nil {
		return
	}

	cidr, _ := vcNet.Mask.Size()
	addr, gatewayAddr, err := vc.GetIp(db, n.VmAdapter.Subnet, n.Virt.Id)
	if err != nil {
		return
	}

	n.InternalAddr = addr
	n.InternalGatewayAddr = gatewayAddr
	n.InternalGatewayAddrCidr = fmt.Sprintf(
		"%s/%d", gatewayAddr.String(), cidr)

	n.InternalAddr6 = vc.GetIp6(addr)
	n.InternalGatewayAddr6 = vc.GetIp6(gatewayAddr)

	n.ExternalMacAddr = vm.GetMacAddrExternal(n.Virt.Id, vc.Id)
	n.InternalMacAddr = vm.GetMacAddrInternal(n.Virt.Id, vc.Id)
	n.HostMacAddr = vm.GetMacAddrHost(n.Virt.Id, vc.Id)

	if n.PhysicalExternalIface6 != n.PhysicalExternalIface {
		n.ExternalMacAddr6 = vm.GetMacAddrExternal6(n.Virt.Id, vc.Id)
	}

	if n.NetworkMode == node.Dhcp {
		n.PhysicalExternalIface = interfaces.GetExternal(
			n.SystemExternalIface)
	} else if n.NetworkMode == node.Static {
		blck, staticAddr, externalIface, e := node.Self.GetStaticAddr(
			db, n.Virt.Id)
		if e != nil {
			err = e
			return
		}

		n.PhysicalExternalIface = externalIface

		staticGateway := blck.GetGateway()
		staticMask := blck.GetMask()
		if staticGateway == nil || staticMask == nil {
			err = &errortypes.ParseError{
				errors.New("qemu: Invalid block gateway cidr"),
			}
			return
		}

		staticSize, _ := staticMask.Size()
		staticCidr := fmt.Sprintf(
			"%s/%d", staticAddr.String(), staticSize)

		n.ExternalAddrCidr = staticCidr
		n.ExternalGatewayAddr = staticGateway
	}

	if n.NetworkMode6 == node.Dhcp {
		if n.NetworkMode == node.Dhcp {
			n.PhysicalExternalIface6 = n.PhysicalExternalIface
		} else {
			n.PhysicalExternalIface6 = interfaces.GetExternal(
				n.SystemExternalIface6)
		}
	} else if n.NetworkMode == node.Static {
		blck, staticAddr, prefix, iface, e := node.Self.GetStaticAddr6(
			db, n.Virt.Id, vc.VpcId)
		if e != nil {
			err = e
			return
		}

		n.PhysicalExternalIface6 = iface

		staticCidr6 := fmt.Sprintf("%s/%d", staticAddr.String(), prefix)
		gateway6 := blck.GetGateway6()

		n.ExternalAddrCidr6 = staticCidr6
		n.ExternalGatewayAddr6 = gateway6
	}

	if n.PhysicalExternalIface6 == "" {
		err = &errortypes.NotFoundError{
			errors.New("qemu: Failed to get external interface6"),
		}
		return
	}

	if n.SpaceExternalIface6 == n.SpaceExternalIface {
		n.ExternalMacAddr6 = n.ExternalMacAddr
	} else {
		n.ExternalMacAddr6 = vm.GetMacAddrExternal6(n.Virt.Id, vc.Id)
	}

	if n.HostNetwork {
		blck, staticAddr, e := node.Self.GetStaticHostAddr(db, n.Virt.Id)
		if e != nil {
			err = e
			return
		}

		n.HostAddr = staticAddr

		hostStaticGateway := blck.GetGateway()
		hostStaticMask := blck.GetMask()
		if hostStaticGateway == nil || hostStaticMask == nil {
			err = &errortypes.ParseError{
				errors.New("qemu: Invalid block gateway cidr"),
			}
			return
		}

		hostStaticSize, _ := hostStaticMask.Size()
		hostStaticCidr := fmt.Sprintf(
			"%s/%d", staticAddr.String(), hostStaticSize)

		n.HostAddrCidr = hostStaticCidr
		n.HostGatewayAddr = hostStaticGateway
	}

	return
}

package netconf

import (
	"fmt"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/interfaces"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/pritunl/pritunl-cloud/vpc"
)

func (n *NetConf) Address(db *database.Database) (err error) {
	vc, err := vpc.Get(db, n.VmAdapter.Vpc)
	if err != nil {
		return
	}

	n.VlanId = vc.VpcId

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
	n.InternalGatewayAddr6 = vc.GetGatewayIp6(addr)
	if err != nil {
		return
	}

	n.ExternalMacAddr = vm.GetMacAddrExternal(n.Virt.Id, vc.Id)
	n.InternalMacAddr = vm.GetMacAddrInternal(n.Virt.Id, vc.Id)
	n.HostMacAddr = vm.GetMacAddrHost(n.Virt.Id, vc.Id)
	n.NodePortMacAddr = vm.GetMacAddrNodePort(n.Virt.Id, vc.Id)

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
	} else if n.NetworkMode6 != node.Disabled &&
		n.NetworkMode6 != node.Oracle {

		n.PhysicalExternalIface = interfaces.GetExternal(
			n.SystemExternalIface)
	}

	if n.NetworkMode6 == node.Static {
		blck, staticAddr, prefix, iface, e := node.Self.GetStaticAddr6(
			db, n.Virt.Id, vc.VpcId, n.PhysicalExternalIface)
		if e != nil {
			err = e
			return
		}

		n.PhysicalExternalIface = iface

		staticCidr6 := fmt.Sprintf("%s/%d", staticAddr.String(), prefix)
		gateway6 := blck.GetGateway6()

		n.ExternalAddrCidr6 = staticCidr6
		n.ExternalGatewayAddr6 = gateway6
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

	if n.NodePortNetwork {
		blck, staticAddr, e := node.Self.GetStaticNodePortAddr(db, n.Virt.Id)
		if e != nil {
			err = e
			return
		}

		n.NodePortAddr = staticAddr

		nodePortStaticMask := blck.GetMask()
		if nodePortStaticMask == nil {
			err = &errortypes.ParseError{
				errors.New("qemu: Invalid block gateway cidr"),
			}
			return
		}

		nodePortStaticSize, _ := nodePortStaticMask.Size()
		nodePortStaticCidr := fmt.Sprintf(
			"%s/%d", staticAddr.String(), nodePortStaticSize)

		n.NodePortAddrCidr = nodePortStaticCidr
	}

	return
}

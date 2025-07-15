package dhcpc

import (
	"encoding/hex"
	"net"

	"github.com/dropbox/godropbox/errors"
	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/insomniacslk/dhcp/dhcpv4/nclient4"
	"github.com/pritunl/pritunl-cloud/imds/server/errortypes"
)

func buildDhLease(lease *Lease, addr net.HardwareAddr) (
	dhLease *nclient4.Lease, err error) {

	xid := TransactionIdUnmarshal(lease.TransactionId)

	offer, err := dhcpv4.New(
		dhcpv4.WithMessageType(dhcpv4.MessageTypeOffer),
		dhcpv4.WithTransactionID(xid),
		dhcpv4.WithHwAddr(addr),
		dhcpv4.WithYourIP(lease.Address.IP),
		dhcpv4.WithServerIP(lease.ServerAddress),
		dhcpv4.WithGatewayIP(lease.Gateway),
		dhcpv4.WithOption(dhcpv4.OptSubnetMask(
			net.IPMask(lease.Address.Mask))),
		dhcpv4.WithOption(dhcpv4.OptRouter(lease.Gateway)),
		dhcpv4.WithOption(dhcpv4.OptServerIdentifier(lease.ServerAddress)),
		dhcpv4.WithOption(dhcpv4.OptIPAddressLeaseTime(lease.LeaseTime)),
	)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "dhcpc: Failed to create offer message"),
		}
		return
	}

	ack, err := dhcpv4.New(
		dhcpv4.WithMessageType(dhcpv4.MessageTypeAck),
		dhcpv4.WithTransactionID(xid),
		dhcpv4.WithHwAddr(addr),
		dhcpv4.WithYourIP(lease.Address.IP),
		dhcpv4.WithServerIP(lease.ServerAddress),
		dhcpv4.WithGatewayIP(lease.Gateway),
		dhcpv4.WithOption(dhcpv4.OptSubnetMask(
			net.IPMask(lease.Address.Mask))),
		dhcpv4.WithOption(dhcpv4.OptRouter(lease.Gateway)),
		dhcpv4.WithOption(dhcpv4.OptServerIdentifier(lease.ServerAddress)),
		dhcpv4.WithOption(dhcpv4.OptIPAddressLeaseTime(lease.LeaseTime)),
	)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "dhcpc: Failed to create ack message"),
		}
		return
	}

	dhLease = &nclient4.Lease{
		Offer: offer,
		ACK:   ack,
	}

	return
}

func extractDhLease(dhLease *nclient4.Lease) (lease *Lease) {
	if dhLease == nil || dhLease.ACK == nil {
		return
	}

	ack := dhLease.ACK

	lease = &Lease{
		Address: &net.IPNet{
			IP:   ack.YourIPAddr,
			Mask: net.IPMask(net.IP{255, 255, 255, 0}),
		},
		Gateway:       ack.GatewayIPAddr,
		ServerAddress: ack.ServerIPAddr,
		TransactionId: ack.TransactionID.String(),
	}

	if subnet := ack.SubnetMask(); subnet != nil {
		lease.Address.Mask = subnet
	}

	if lease.Gateway.Equal(net.IPv4zero) || lease.Gateway == nil {
		if routers := ack.Router(); len(routers) > 0 {
			lease.Gateway = routers[0]
		}
	}

	serverID := ack.ServerIdentifier()
	if serverID != nil {
		lease.ServerAddress = serverID
	}

	leaseTime := ack.IPAddressLeaseTime(0)
	if leaseTime > 0 {
		lease.LeaseTime = leaseTime
	}

	return
}

func TransactionIdUnmarshal(str string) dhcpv4.TransactionID {
	var tranId dhcpv4.TransactionID

	if len(str) >= 2 && str[:2] == "0x" {
		str = str[2:]
	}

	bytes, err := hex.DecodeString(str)
	if err != nil {
		return tranId
	}

	if len(bytes) != 4 {
		return tranId
	}

	copy(tranId[:], bytes)
	return tranId
}

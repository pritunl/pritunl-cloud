package dhcpc

import (
	"context"
	"net"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/insomniacslk/dhcp/dhcpv4/nclient4"
	"github.com/pritunl/pritunl-cloud/imds/server/errortypes"
)

type Lease struct {
	Address       *net.IPNet
	Gateway       net.IP
	ServerAddress net.IP
	LeaseTime     time.Duration
	TransactionId string
}

func (l *Lease) Renew() (ok bool, err error) {
	iface, err := net.InterfaceByName(DhcpIface)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "dhcpc: Failed to find interface"),
		}
		return
	}

	dhLease, err := buildDhLease(l, iface.HardwareAddr)
	if err != nil {
		return
	}

	serverAddr := &net.UDPAddr{
		IP:   l.ServerAddress,
		Port: nclient4.ServerPort,
	}

	client, err := nclient4.New(DhcpIface,
		nclient4.WithServerAddr(serverAddr),
		nclient4.WithTimeout(DhcpTimeout),
		nclient4.WithRetry(DhcpRetries),
		nclient4.WithUnicast(&net.UDPAddr{
			IP:   l.Address.IP,
			Port: nclient4.ClientPort,
		}),
	)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "dhcpc: Failed to create client"),
		}
		return
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), DhcpTimeout)
	defer cancel()

	renewedLease, err := client.Renew(ctx, dhLease,
		dhcpv4.WithOption(dhcpv4.OptMaxMessageSize(MaxMessageSize)),
	)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "dhcpc: Failed to exchange renewal"),
		}
		return
	}

	renewed := extractDhLease(renewedLease)
	if renewed != nil {
		ok = true
		l.Address = renewed.Address
		l.Gateway = renewed.Gateway
		l.ServerAddress = renewed.ServerAddress
		l.LeaseTime = renewed.LeaseTime
		l.TransactionId = renewed.TransactionId
	}

	return
}

func (l *Lease) Exchange() (ok bool, err error) {
	client, err := nclient4.New(
		DhcpIface,
		nclient4.WithTimeout(DhcpTimeout),
		nclient4.WithRetry(DhcpRetries),
	)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "dhcpc: Failed to create client"),
		}
		return
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), DhcpTimeout)
	defer cancel()

	nclientLease, err := client.Request(
		ctx,
		dhcpv4.WithOption(dhcpv4.OptMaxMessageSize(MaxMessageSize)),
	)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "dhcpc: IPv4 exchange failed"),
		}
		return
	}

	lease := extractDhLease(nclientLease)
	if lease != nil {
		ok = true
		l.Address = lease.Address
		l.Gateway = lease.Gateway
		l.ServerAddress = lease.ServerAddress
		l.LeaseTime = lease.LeaseTime
		l.TransactionId = lease.TransactionId
	}

	return
}

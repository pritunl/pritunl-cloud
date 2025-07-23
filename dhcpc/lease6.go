package dhcpc

import (
	"context"
	"net"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/insomniacslk/dhcp/dhcpv6"
	"github.com/insomniacslk/dhcp/dhcpv6/nclient6"
	"github.com/insomniacslk/dhcp/iana"
	"github.com/pritunl/pritunl-cloud/imds/server/errortypes"
)

func (l *Lease) Renew6() (ok bool, err error) {
	if l.Address6 == nil || l.Address6.IP == nil || l.ServerAddress == nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "dhcpc: Cannot call renew with unset address"),
		}
		return
	}

	iface, err := net.InterfaceByName(l.Iface)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "dhcpc: Failed to find interface"),
		}
		return
	}

	client, err := nclient6.New(l.Iface,
		nclient6.WithTimeout(DhcpTimeout),
		nclient6.WithRetry(DhcpRetries),
	)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "dhcpc: Failed to create dhcp6 client"),
		}
		return
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), DhcpTimeout)
	defer cancel()

	serverAddr := &net.UDPAddr{
		IP:   l.ServerAddress6,
		Port: 547,
	}

	iaAddr := &dhcpv6.OptIAAddress{
		IPv6Addr:          l.Address6.IP,
		PreferredLifetime: PreferredTtl,
		ValidLifetime:     PreferredTtl,
	}

	msg, err := dhcpv6.NewMessage()
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "dhcpc: Failed to create dhcp6 message"),
		}
		return
	}
	msg.MessageType = dhcpv6.MessageTypeRenew
	msg.AddOption(dhcpv6.OptClientID(&dhcpv6.DUIDLLT{
		HWType:        iana.HWTypeEthernet,
		Time:          dhcpv6.GetTime(),
		LinkLayerAddr: iface.HardwareAddr,
	}))
	msg.AddOption(dhcpv6.OptServerID(l.ServerId6))
	msg.AddOption(&dhcpv6.OptIANA{
		IaId: l.IaId6,
		Options: dhcpv6.IdentityOptions{
			Options: []dhcpv6.Option{iaAddr},
		},
	})
	// msg.AddOption(dhcpv6.OptRequestedOption(
	// 	dhcpv6.OptionDNSRecursiveNameServer,
	// 	dhcpv6.OptionDomainSearchList,
	// ))
	//msg.AddOption(dhcpv6.OptElapsedTime(0))

	reply, err := client.SendAndRead(
		ctx,
		serverAddr,
		msg,
		nclient6.IsMessageType(dhcpv6.MessageTypeRenew),
	)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "dhcpc: Failed to renew DHCPv6 lease"),
		}
		return
	}

	renewed := extractDhcpv6Lease(reply, l.Iface)
	if renewed != nil && renewed.Address6 != nil {
		ok = true
		l.Address6 = renewed.Address6
		l.ServerAddress6 = renewed.ServerAddress6
		l.LeaseTime6 = renewed.LeaseTime6
		l.TransactionId6 = renewed.TransactionId6
		l.ServerId6 = renewed.ServerId6
	}

	return
}

func (l *Lease) Exchange6() (ok bool, err error) {
	iface, err := net.InterfaceByName(l.Iface)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "dhcpc: Failed to find interface"),
		}
		return
	}

	client, err := nclient6.New(
		l.Iface,
		nclient6.WithTimeout(DhcpTimeout),
		nclient6.WithRetry(DhcpRetries),
	)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "dhcpc: Failed to create DHCPv6 client"),
		}
		return
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), DhcpTimeout)
	defer cancel()

	l.IaId6 = [4]byte{0, 0, byte(iface.Index >> 8), byte(iface.Index)}

	modifiers := []dhcpv6.Modifier{
		dhcpv6.WithClientID(&dhcpv6.DUIDLLT{
			HWType:        iana.HWTypeEthernet,
			Time:          dhcpv6.GetTime(),
			LinkLayerAddr: iface.HardwareAddr,
		}),
		dhcpv6.WithRequestedOptions(
			dhcpv6.OptionDNSRecursiveNameServer,
			dhcpv6.OptionDomainSearchList,
		),
		dhcpv6.WithFQDN(0, ""),
		dhcpv6.WithIAID(l.IaId6),
	}

	if l.Address6 != nil && l.Address6.IP != nil {
		iaAddr := &dhcpv6.OptIAAddress{
			IPv6Addr:          l.Address6.IP,
			PreferredLifetime: 3600,
			ValidLifetime:     7200,
		}
		iaNa := &dhcpv6.OptIANA{
			IaId: l.IaId6,
			Options: dhcpv6.IdentityOptions{
				Options: []dhcpv6.Option{iaAddr},
			},
		}
		modifiers = append(modifiers, dhcpv6.WithOption(iaNa))
	}

	reply, err := client.RapidSolicit(ctx, modifiers...)
	if err != nil {
		reply, err = client.Solicit(ctx, modifiers...)
		if err != nil {
			err = &errortypes.RequestError{
				errors.Wrap(err, "dhcpc: DHCPv6 exchange failed"),
			}
			return
		}
	}

	lease := extractDhcpv6Lease(reply, l.Iface)
	if lease != nil && lease.Address6 != nil {
		ok = true
		l.Address6 = lease.Address6
		l.ServerAddress6 = lease.ServerAddress6
		l.LeaseTime6 = lease.LeaseTime6
		l.TransactionId6 = lease.TransactionId6
		l.ServerId6 = lease.ServerId6
	}

	return
}

func extractDhcpv6Lease(reply *dhcpv6.Message, ifaceName string) *Lease {
	if reply == nil {
		return nil
	}

	lease := &Lease{
		Iface:          ifaceName,
		TransactionId6: reply.TransactionID.String(),
	}

	serverID := reply.Options.ServerID()
	if serverID != nil {
		lease.ServerId6 = reply.Options.ServerID()
	}

	// // Extract server address from relay message or use link-local
	// reply.Options.RelayMessage
	// if relayMsg := reply.GetOneOption(dhcpv6.OptionRelayMsg); relayMsg != nil {
	// 	// Server address might be in relay message
	// 	if rm, ok := relayMsg.(*dhcpv6.OptRelayMessage); ok && rm.RelayMessage != nil {
	// 		lease.ServerAddress = rm.RelayMessage.PeerAddr
	// 	}
	// }

	if lease.ServerAddress6 == nil {
		lease.ServerAddress6 = dhcpv6.AllDHCPRelayAgentsAndServers
	}

	iana := reply.Options.OneIANA()
	if iana != nil {
		lease.IaId6 = iana.IaId

		for _, opt := range iana.Options.Options {
			if addr, ok := opt.(*dhcpv6.OptIAAddress); ok {
				lease.Address6 = &net.IPNet{
					IP:   addr.IPv6Addr,
					Mask: net.CIDRMask(64, 128),
				}
				lease.LeaseTime6 = time.Duration(
					addr.ValidLifetime) * time.Second
				break
			}
		}
	}

	return lease
}

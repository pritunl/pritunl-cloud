package dhcps

import (
	"fmt"
	"net"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/insomniacslk/dhcp/dhcpv6"
	"github.com/insomniacslk/dhcp/dhcpv6/server6"
	"github.com/insomniacslk/dhcp/iana"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/sirupsen/logrus"
)

type Server6 struct {
	Iface        string   `json:"iface"`
	ClientIp     string   `json:"client_ip"`
	GatewayIp    string   `json:"gateway_ip"`
	PrefixLen    int      `json:"prefix_len"`
	DnsServers   []string `json:"dns_servers"`
	Mtu          int      `json:"mtu"`
	Lifetime     int      `json:"lifetime"`
	Debug        bool     `json:"debug"`
	serverId     dhcpv6.Duid
	dnsServersIp []net.IP
	server       *server6.Server
	lifetime     time.Duration
}

func (s *Server6) handler(conn net.PacketConn, peer net.Addr,
	req dhcpv6.DHCPv6) {

	err := s.handleMsg(conn, peer, req)
	if err != nil {
		if s.Debug {
			logrus.WithFields(logrus.Fields{
				"peer":  peer.String(),
				"error": err,
			}).Error("dhcps: DHCPv6 handler error")
		}
	}
}

func (s *Server6) handleMsg(conn net.PacketConn, peer net.Addr,
	req dhcpv6.DHCPv6) (err error) {

	msg, err := req.GetInnerMessage()
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "dhcps: DHCPv6 get inner message error"),
		}
		return
	}

	clientId := msg.Options.ClientID()
	if clientId == nil {
		err = &errortypes.ParseError{
			errors.New("dhcps: DHCPv6 missing client id"),
		}
		return
	}

	serverId := msg.Options.ServerID()

	if s.Debug {
		fmt.Printf("Peer: %s\n", peer.String())
		fmt.Println(msg.Summary())
	}

	switch msg.Type() {
	case dhcpv6.MessageTypeSolicit, dhcpv6.MessageTypeConfirm,
		dhcpv6.MessageTypeRebind:

		if serverId != nil {
			err = &errortypes.ParseError{
				errors.New("dhcps: DHCPv6 invalid server id"),
			}
			return
		}
	case dhcpv6.MessageTypeRequest, dhcpv6.MessageTypeRenew,
		dhcpv6.MessageTypeRelease, dhcpv6.MessageTypeDecline:

		if serverId == nil {
			err = &errortypes.ParseError{
				errors.New("dhcps: DHCPv6 missing server id"),
			}
			return
		}

		if !serverId.Equal(s.serverId) {
			err = &errortypes.ParseError{
				errors.New("dhcps: DHCPv6 server id mismatch"),
			}
			return
		}
	}

	var resp dhcpv6.DHCPv6
	switch msg.Type() {
	case dhcpv6.MessageTypeSolicit:
		rapidCommit := msg.GetOneOption(dhcpv6.OptionRapidCommit)
		if rapidCommit != nil {
			resp, err = dhcpv6.NewReplyFromMessage(msg)
			if err != nil {
				err = &errortypes.ParseError{
					errors.Wrap(err, "dhcps: DHCPv6 new reply "+
						"from message error"),
				}
				return
			}
		} else {
			resp, err = dhcpv6.NewAdvertiseFromSolicit(msg)
			if err != nil {
				err = &errortypes.ParseError{
					errors.Wrap(err, "dhcps: DHCPv6 new advertise "+
						"from solicit error"),
				}
				return
			}
		}
		break
	case dhcpv6.MessageTypeRequest, dhcpv6.MessageTypeConfirm,
		dhcpv6.MessageTypeRenew, dhcpv6.MessageTypeRebind,
		dhcpv6.MessageTypeRelease, dhcpv6.MessageTypeInformationRequest:

		resp, err = dhcpv6.NewReplyFromMessage(msg)
		if err != nil {
			err = &errortypes.ParseError{
				errors.Wrap(err, "dhcps: DHCPv6 new reply "+
					"from message error"),
			}
			return
		}
		break
	default:
		err = &errortypes.ParseError{
			errors.New("dhcps: Unknown DHCPv6 message type"),
		}
		return
	}

	resp.AddOption(dhcpv6.OptServerID(s.serverId))

	err = s.process(msg, req, resp)
	if err != nil {
		return
	}

	if s.Debug {
		fmt.Printf("Peer: %s\n", peer.String())
		fmt.Println(resp.Summary())
	}

	_, err = conn.WriteTo(resp.ToBytes(), peer)
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "dhcps: DHCPv6 reply write error"),
		}
		return
	}

	return
}

func (s *Server6) process(msg *dhcpv6.Message,
	req, resp dhcpv6.DHCPv6) (err error) {

	switch msg.Type() {
	case dhcpv6.MessageTypeSolicit, dhcpv6.MessageTypeRequest,
		dhcpv6.MessageTypeConfirm, dhcpv6.MessageTypeRenew,
		dhcpv6.MessageTypeRebind:

		break
	default:
		err = &errortypes.ParseError{
			errors.Newf("dhcps: DHCPv6 ignore message type %s", msg.Type()),
		}
		return
	}

	oia := &dhcpv6.OptIANA{
		T1: s.lifetime / 2,
		T2: time.Duration(float32(s.lifetime) / 1.5),
	}

	roia := msg.Options.OneIANA()
	if roia != nil {
		copy(oia.IaId[:], roia.IaId[:])
	} else {
		copy(oia.IaId[:], []byte("CLOUD"))
	}

	oiaAddr := &dhcpv6.OptIAAddress{
		IPv6Addr:          net.ParseIP(s.ClientIp),
		PreferredLifetime: s.lifetime,
		ValidLifetime:     s.lifetime,
	}

	oia.Options = dhcpv6.IdentityOptions{
		Options: []dhcpv6.Option{
			oiaAddr,
		},
	}

	resp.AddOption(oia)

	if msg.IsOptionRequested(dhcpv6.OptionDNSRecursiveNameServer) &&
		s.dnsServersIp != nil {

		resp.UpdateOption(dhcpv6.OptDNS(s.dnsServersIp...))
	}

	fqdn := msg.GetOneOption(dhcpv6.OptionFQDN)
	if fqdn != nil {
		resp.AddOption(fqdn)
	}

	resp.AddOption(&dhcpv6.OptStatusCode{
		StatusCode:    iana.StatusSuccess,
		StatusMessage: "success",
	})

	return
}

func (s *Server6) Start() (err error) {
	s.lifetime = time.Duration(s.Lifetime) * time.Second

	iface, err := net.InterfaceByName(s.Iface)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "dhcps: Failed to find network interface"),
		}
		return
	}

	if s.DnsServers != nil && len(s.DnsServers) > 0 {
		dnsServers := []net.IP{}
		for _, dnsServer := range s.DnsServers {
			dnsServers = append(dnsServers, net.ParseIP(dnsServer))
		}
		s.dnsServersIp = dnsServers
	}

	s.serverId = dhcpv6.Duid{
		Type:          dhcpv6.DUID_LLT,
		HwType:        iana.HWTypeEthernet,
		LinkLayerAddr: iface.HardwareAddr,
		Time:          dhcpv6.GetTime(),
	}

	host6 := &net.UDPAddr{
		IP:   net.ParseIP("::"),
		Port: dhcpv6.DefaultServerPort,
	}

	s.server, err = server6.NewServer(iface.Name, host6, s.handler)
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "dhcps: Failed to create server6"),
		}
		return
	}

	err = s.server.Serve()
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "dhcps: Failed to start server6"),
		}
		return
	}

	return
}

package dhcps

import (
	"fmt"
	"net"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/insomniacslk/dhcp/dhcpv4/server4"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/sirupsen/logrus"
)

type Server4 struct {
	Iface        string   `json:"iface"`
	ClientIp     string   `json:"client_ip"`
	GatewayIp    string   `json:"gateway_ip"`
	PrefixLen    int      `json:"prefix_len"`
	DnsServers   []string `json:"dns_servers"`
	Mtu          int      `json:"mtu"`
	Lifetime     int      `json:"lifetime"`
	Debug        bool     `json:"debug"`
	dnsServersIp []net.IP
	server       *server4.Server
	lifetime     time.Duration
}

func (s *Server4) handler(conn net.PacketConn, peer net.Addr,
	req *dhcpv4.DHCPv4) {

	err := s.handleMsg(conn, peer, req)
	if err != nil {
		if s.Debug {
			logrus.WithFields(logrus.Fields{
				"peer":  peer.String(),
				"error": err,
			}).Error("dhcps: DHCPv4 handler error")
		}
	}
}

func (s *Server4) handleMsg(conn net.PacketConn, peer net.Addr,
	req *dhcpv4.DHCPv4) (err error) {

	if req.MessageType() != dhcpv4.MessageTypeDiscover &&
		req.MessageType() != dhcpv4.MessageTypeRequest {

		return
	}

	if s.Debug {
		fmt.Printf("Peer: %s\n", peer.String())
		fmt.Println(req.Summary())
	}

	resp, err := dhcpv4.NewReplyFromRequest(req)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "dhcps: Failed to create reply"),
		}
		return
	}

	if req.MessageType() == dhcpv4.MessageTypeRequest {
		resp.UpdateOption(dhcpv4.OptMessageType(dhcpv4.MessageTypeAck))
	} else {
		resp.UpdateOption(dhcpv4.OptMessageType(dhcpv4.MessageTypeOffer))
	}

	gatewayIp := net.ParseIP(s.GatewayIp)
	clientIp := net.ParseIP(s.ClientIp)

	resp.YourIPAddr = clientIp
	resp.UpdateOption(dhcpv4.OptRouter(gatewayIp))
	resp.UpdateOption(dhcpv4.OptSubnetMask(
		net.CIDRMask(s.PrefixLen, net.IPv4len*8)))
	resp.UpdateOption(dhcpv4.OptServerIdentifier(gatewayIp))
	resp.UpdateOption(dhcpv4.OptIPAddressLeaseTime(s.lifetime))

	requested := req.ParameterRequestList()
	if requested.Has(dhcpv4.OptionDomainNameServer) {
		resp.UpdateOption(dhcpv4.OptDNS(s.dnsServersIp...))
	}

	if s.Mtu != 0 {
		resp.UpdateOption(dhcpv4.Option{
			Code:  dhcpv4.OptionInterfaceMTU,
			Value: dhcpv4.Uint16(s.Mtu),
		})
	}

	if s.Debug {
		fmt.Printf("Peer: %s\n", peer.String())
		fmt.Println(resp.Summary())
	}

	_, err = conn.WriteTo(resp.ToBytes(), peer)
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "dhcps: DHCPv4 resp write error"),
		}
		return
	}

	return
}

func (s *Server4) Start() (err error) {
	logrus.WithFields(logrus.Fields{
		"iface":       s.Iface,
		"client_ip":   s.ClientIp,
		"gateway_ip":  s.GatewayIp,
		"prefix_len":  s.PrefixLen,
		"dns_servers": s.DnsServers,
		"mtu":         s.Mtu,
		"lifetime":    s.Lifetime,
		"debug":       s.Debug,
	}).Info("dhcps: Starting server4")

	s.lifetime = time.Duration(s.Lifetime) * time.Second

	if s.DnsServers != nil && len(s.DnsServers) > 0 {
		dnsServers := []net.IP{}
		for _, dnsServer := range s.DnsServers {
			dnsServers = append(dnsServers, net.ParseIP(dnsServer))
		}
		s.dnsServersIp = dnsServers
	}

	host4 := &net.UDPAddr{
		Port: dhcpv4.ServerPort,
	}

	s.server, err = server4.NewServer(s.Iface, host4, s.handler)
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "dhcps: Failed to create server4"),
		}
		return
	}

	err = s.server.Serve()
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "dhcps: Failed to start server4"),
		}
		return
	}

	return
}

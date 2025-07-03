package dhcps

import (
	"fmt"
	"net"
	"net/netip"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/mdlayher/ndp"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/sirupsen/logrus"
)

type ServerNdp struct {
	Iface       string   `json:"iface"`
	ClientIp    string   `json:"client_ip"`
	GatewayIp   string   `json:"gateway_ip"`
	PrefixLen   int      `json:"prefix_len"`
	DnsServers  []string `json:"dns_servers"`
	Mtu         int      `json:"mtu"`
	Lifetime    int      `json:"lifetime"`
	Delay       int      `json:"delay"`
	Debug       bool     `json:"debug"`
	iface       *net.Interface
	gatewayAddr netip.Addr
	prefixAddr  netip.Addr
	lifetime    time.Duration
	delay       time.Duration
}

func (s *ServerNdp) Start() (err error) {
	logrus.WithFields(logrus.Fields{
		"iface":       s.Iface,
		"client_ip":   s.ClientIp,
		"gateway_ip":  s.GatewayIp,
		"prefix_len":  s.PrefixLen,
		"dns_servers": s.DnsServers,
		"mtu":         s.Mtu,
		"lifetime":    s.Lifetime,
		"delay":       s.Delay,
		"debug":       s.Debug,
	}).Info("dhcps: Starting ndp server")

	s.lifetime = time.Duration(s.Lifetime) * time.Second
	s.delay = time.Duration(s.Delay) * time.Second

	s.iface, err = net.InterfaceByName(s.Iface)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "dhcps: Failed to find network interface"),
		}
		return
	}

	s.gatewayAddr, err = netip.ParseAddr(s.GatewayIp)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "dhcps: Failed to parse gateway addr"),
		}
		return
	}

	_, prefixNet, err := net.ParseCIDR(fmt.Sprintf(
		"%s/%d", s.ClientIp, s.PrefixLen))
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "dhcps: Failed to parse client IP and prefix"),
		}
		return
	}

	prefix, err := netip.ParsePrefix(prefixNet.String())
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "dhcps: Failed to parse client addr and prefix"),
		}
		return
	}

	s.prefixAddr = prefix.Masked().Addr()

	for {
		err = s.run()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("dhcps: NDP server error")
		}

		time.Sleep(s.delay)
	}
}

func (s *ServerNdp) run() (err error) {
	conn, _, err := ndp.Listen(s.iface, ndp.LinkLocal)
	if err != nil {
		err = &errortypes.NetworkError{
			errors.Wrap(err, "dhcps: Failed to listen for NDP messages"),
		}
		return
	}
	defer conn.Close()

	err = conn.JoinGroup(netip.MustParseAddr("ff02::2"))
	if err != nil {
		err = &errortypes.NetworkError{
			errors.Wrap(err, "dhcps: Failed to join NDP group"),
		}
		return
	}

	err = s.sendAdvertise(conn, netip.IPv6LinkLocalAllNodes())
	if err != nil {
		return
	}

	return
}

func (s *ServerNdp) sendAdvertise(conn *ndp.Conn, dst netip.Addr) (err error) {
	opts := []ndp.Option{
		&ndp.PrefixInformation{
			Prefix:                         s.prefixAddr,
			PrefixLength:                   uint8(s.PrefixLen),
			OnLink:                         true,
			AutonomousAddressConfiguration: false,
			ValidLifetime:                  s.lifetime,
			PreferredLifetime:              s.lifetime,
		},
		&ndp.LinkLayerAddress{
			Direction: ndp.Source,
			Addr:      s.iface.HardwareAddr,
		},
	}

	if s.Mtu != 0 {
		opts = append(opts, &ndp.MTU{
			MTU: uint32(s.Mtu),
		})
	}

	if len(s.DnsServers) > 0 {
		dnsAddrs := make([]netip.Addr, 0, len(s.DnsServers))
		for _, dns := range s.DnsServers {
			addr, err := netip.ParseAddr(dns)
			if err == nil && addr.Is6() {
				dnsAddrs = append(dnsAddrs, addr)
			}
		}
		if len(dnsAddrs) > 0 {
			opts = append(opts, &ndp.RecursiveDNSServer{
				Lifetime: s.lifetime,
				Servers:  dnsAddrs,
			})
		}
	}

	msgRa := &ndp.RouterAdvertisement{
		CurrentHopLimit:           64,
		RouterSelectionPreference: ndp.Medium,
		RouterLifetime:            s.lifetime,
		ManagedConfiguration:      true,
		OtherConfiguration:        true,
		Options:                   opts,
	}

	if s.Debug {
		logrus.WithFields(logrus.Fields{
			"gateway":         s.gatewayAddr.String(),
			"prefix":          s.prefixAddr.String(),
			"prefix_len":      s.PrefixLen,
			"router_lifetime": s.lifetime,
		}).Info("dhcps: Sending router advertisement")
	}

	err = conn.WriteTo(msgRa, nil, dst)
	if err != nil {
		err = &errortypes.NetworkError{
			errors.Wrap(err, "dhcps: Failed to write NDP message"),
		}
		return
	}

	return
}

func (s *ServerNdp) readSolicitations(conn *ndp.Conn) (err error) {
	if s.Debug {
		logrus.WithFields(logrus.Fields{
			"gateway":         s.gatewayAddr.String(),
			"prefix":          s.prefixAddr.String(),
			"prefix_len":      s.PrefixLen,
			"router_lifetime": s.lifetime,
		}).Info("dhcps: Reading router solicitations")
	}

	err = conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "dhcps: Failed to set deadline"),
		}
		return
	}

	msg, _, from, err := conn.ReadFrom()
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			err = nil
			return
		}

		err = &errortypes.ReadError{
			errors.Wrap(err, "dhcps: Failed to read NDP message"),
		}
		return
	}

	if _, ok := msg.(*ndp.RouterSolicitation); ok {
		if s.Debug {
			logrus.WithFields(logrus.Fields{
				"from": from.String(),
			}).Info("dhcps: Received Router Solicitation")
		}

		err = s.sendAdvertise(conn, from)
		if err != nil {
			return
		}
	}

	return
}

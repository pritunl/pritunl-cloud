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

	prefix, err := netip.ParsePrefix(
		fmt.Sprintf("%s/%d", s.ClientIp, s.PrefixLen))
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "dhcps: Failed to parse client addr and prefix"),
		}
		return
	}

	s.prefixAddr = prefix.Masked().Addr()

	prefix.Addr()

	for {
		err = s.run()
		if err != nil {
			if s.Debug {
				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Error("dhcps: NDP server error")
			}
		}

		time.Sleep(s.delay)
	}
}

func (s *ServerNdp) run() (err error) {
	conn, _, err := ndp.Listen(s.iface, ndp.LinkLocal)
	if err != nil {
		err = &errortypes.NetworkError{
			errors.Wrap(err, "dhcps: Failed to write NDP message"),
		}
		return
	}
	defer func() {
		_ = conn.Close()
	}()

	err = conn.SetDeadline(time.Now().Add(10 * time.Second))
	if err != nil {
		err = &errortypes.NetworkError{
			errors.Wrap(err, "dhcps: Failed to set deadline"),
		}
		return
	}

	opts := []ndp.Option{
		&ndp.PrefixInformation{
			Prefix:                         s.prefixAddr,
			PrefixLength:                   uint8(s.PrefixLen),
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

	msgRa := &ndp.RouterAdvertisement{
		CurrentHopLimit:           64,
		RouterSelectionPreference: ndp.Medium,
		RouterLifetime:            30 * time.Second,
		ManagedConfiguration:      true,
		OtherConfiguration:        true,
		Options:                   opts,
	}

	err = conn.JoinGroup(netip.MustParseAddr("ff02::2"))
	if err != nil {
		err = &errortypes.NetworkError{
			errors.Wrap(err, "dhcps: Failed to join NDP group"),
		}
		return
	}

	if s.Debug {
		fmt.Println("Send RA")
	}

	err = conn.WriteTo(msgRa, nil, netip.IPv6LinkLocalAllNodes())
	if err != nil {
		err = &errortypes.NetworkError{
			errors.Wrap(err, "dhcps: Failed to write NDP message"),
		}
		return
	}

	return
}

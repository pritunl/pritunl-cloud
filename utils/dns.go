package utils

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

func DnsLookup(server, host string) (addrs []string, err error) {
	serverIp := net.ParseIP(server)
	if serverIp == nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "utils: Invalid DNS server address"),
		}
		return
	}

	if serverIp.To4() == nil {
		server = fmt.Sprintf("[%s]:53", serverIp.String())
	} else {
		server = fmt.Sprintf("%s:53", serverIp.String())
	}

	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network,
			address string) (net.Conn, error) {

			dialer := net.Dialer{
				Timeout: 3 * time.Second,
			}

			return dialer.DialContext(ctx, network, server)
		},
	}

	addrs, err = resolver.LookupHost(context.Background(), host)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "utils: DNS lookup failed"),
		}
		return
	}

	if addrs == nil {
		addrs = []string{}
	}

	return
}

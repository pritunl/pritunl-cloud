package proxy

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/imds/server/errortypes"
)

var (
	ResolverLock    sync.RWMutex
	IpDatabase      = set.NewSet()
	ResolverCache   = map[string]*Remote{}
	HostNetwork     *net.IPNet
	NodePortNetwork *net.IPNet
)

type Remote struct {
	Timestamp time.Time
	Remote    net.IP
}

func ResolverValidate(ip net.IP) bool {
	if IpDatabase.Contains(ip.String()) {
		return true
	}
	if HostNetwork.Contains(ip) {
		return true
	}
	if NodePortNetwork.Contains(ip) {
		return true
	}
	return false
}

func Resolve(hostname string) (remote net.IP, err error) {
	defer func() {
		if remote != nil {
			fmt.Println(hostname, remote.String(), err)
		} else {
			fmt.Println(hostname, err)
		}
	}()

	ResolverLock.RLock()
	cached, ok := ResolverCache[hostname]
	if ok {
		ResolverLock.RUnlock()

		remote = cached.Remote
		return
	}
	ResolverLock.RUnlock()

	ip := net.ParseIP(hostname)
	if ip != nil {
		ResolverLock.RLock()
		contains := ResolverValidate(ip)
		ResolverLock.RUnlock()

		if !contains {
			err = &errortypes.RequestError{
				errors.New("proxy: Balancer resolved address not in database"),
			}
			return
		}

		remote = ip
	} else {
		ips, e := net.LookupIP(hostname)
		if e != nil {
			err = &errortypes.RequestError{
				errors.Wrap(e, "proxy: Balancer resolve error"),
			}
			return
		}

		ResolverLock.RLock()
		for _, ip := range ips {
			if ResolverValidate(ip) {
				remote = ip
				break
			}
		}
		ResolverLock.RUnlock()

		if remote == nil {
			err = &errortypes.RequestError{
				errors.New("proxy: Balancer resolved address not in database"),
			}
			return
		}
	}

	ResolverLock.Lock()
	ResolverCache[hostname] = &Remote{
		Timestamp: time.Now(),
		Remote:    remote,
	}
	ResolverLock.Unlock()

	return
}

type StaticDialer struct {
	dialer *net.Dialer
}

func (d *StaticDialer) DialContext(ctx context.Context, network, addr string) (
	conn net.Conn, err error) {

	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	remote, err := Resolve(host)
	if err != nil {
		return
	}

	conn, err = d.dialer.DialContext(
		ctx, network, net.JoinHostPort(remote.String(), port))
	if err == nil {
		return
	}

	return
}

func NewStaticDialer(dialer *net.Dialer) *StaticDialer {
	return &StaticDialer{
		dialer: dialer,
	}
}

func init() {
	_, HostNetwork, _ = net.ParseCIDR("0.0.0.0/32")
	_, NodePortNetwork, _ = net.ParseCIDR("0.0.0.0/32")
}

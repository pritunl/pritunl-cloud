package proxy

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/imds/server/errortypes"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/settings"
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

func ResolverRefresh(db *database.Database) (err error) {
	coll := db.Instances()
	ipDatabase := set.NewSet()
	ttl := time.Duration(settings.Router.ProxyResolverTtl) * time.Second

	cursor, err := coll.Find(
		db,
		&bson.M{},
		&options.FindOptions{
			Projection: &bson.M{
				"public_ips":         1,
				"public_ips6":        1,
				"oracle_private_ips": 1,
				"oracle_public_ips":  1,
				"oracle_public_ips6": 1,
			},
		},
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		inst := &instance.Instance{}

		err = cursor.Decode(inst)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		for _, ipStr := range inst.PublicIps {
			ip := net.ParseIP(ipStr)
			if ip != nil {
				ipDatabase.Add(ip.String())
			}
		}

		for _, ipStr := range inst.PublicIps6 {
			ip := net.ParseIP(ipStr)
			if ip != nil {
				ipDatabase.Add(ip.String())
			}
		}

		for _, ipStr := range inst.OraclePublicIps {
			ip := net.ParseIP(ipStr)
			if ip != nil {
				ipDatabase.Add(ip.String())
			}
		}

		for _, ipStr := range inst.OraclePublicIps6 {
			ip := net.ParseIP(ipStr)
			if ip != nil {
				ipDatabase.Add(ip.String())
			}
		}

		for _, ipStr := range inst.OraclePrivateIps {
			ip := net.ParseIP(ipStr)
			if ip != nil {
				ipDatabase.Add(ip.String())
			}
		}
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	_, hostNetwork, err := net.ParseCIDR(
		settings.Hypervisor.HostNetwork)
	if err != nil {
		return
	}

	_, nodePortNetwork, err := net.ParseCIDR(
		settings.Hypervisor.NodePortNetwork)
	if err != nil {
		return
	}

	now := time.Now()
	ResolverLock.Lock()
	for hostname, cached := range ResolverCache {
		if now.Sub(cached.Timestamp) > ttl {
			delete(ResolverCache, hostname)
		}
	}
	IpDatabase = ipDatabase
	HostNetwork = hostNetwork
	NodePortNetwork = nodePortNetwork
	ResolverLock.Unlock()

	return
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

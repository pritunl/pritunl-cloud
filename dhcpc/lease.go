package dhcpc

import (
	"net"
	"sync"
	"time"

	"github.com/insomniacslk/dhcp/dhcpv6"
)

type Lease struct {
	Timestamp          time.Time     `json:"timestamp"`
	Iface              string        `json:"iface"`
	Iface6             string        `json:"iface6"`
	Address            *net.IPNet    `json:"address"`
	Gateway            net.IP        `json:"gateway"`
	Address6           *net.IPNet    `json:"address6"`
	ServerAddress      net.IP        `json:"server"`
	ServerAddress6     net.IP        `json:"server6"`
	LeaseTime          time.Duration `json:"ttl"`
	LeaseTime6         time.Duration `json:"ttl6"`
	PreferredLifetime6 time.Duration `json:"-"`
	ValidLifetime6     time.Duration `json:"-"`
	TransactionId      string        `json:"-"`
	TransactionId6     string        `json:"-"`
	IaId6              [4]byte       `json:"-"`
	ServerId6          dhcpv6.DUID   `json:"-"`
	progress4          time.Time
	progress6          time.Time
	lock               sync.Mutex
}

func (l *Lease) touch4() {
	l.lock.Lock()
	l.progress4 = time.Now()
	l.lock.Unlock()
}

func (l *Lease) touch6() {
	l.lock.Lock()
	l.progress6 = time.Now()
	l.lock.Unlock()
}

func (l *Lease) heartbeat(ip4, ip6 bool) (timestamp time.Time) {
	l.lock.Lock()
	defer l.lock.Unlock()

	if ip4 {
		timestamp = l.progress4
	}
	if ip6 && (timestamp.IsZero() || l.progress6.Before(timestamp)) {
		timestamp = l.progress6
	}
	if timestamp.IsZero() {
		timestamp = time.Now()
	}

	return
}

func (l *Lease) Copy() *Lease {
	l.lock.Lock()
	defer l.lock.Unlock()

	return &Lease{
		Timestamp:          l.Timestamp,
		Iface:              l.Iface,
		Iface6:             l.Iface6,
		Address:            l.Address,
		Gateway:            l.Gateway,
		Address6:           l.Address6,
		ServerAddress:      l.ServerAddress,
		ServerAddress6:     l.ServerAddress6,
		LeaseTime:          l.LeaseTime,
		LeaseTime6:         l.LeaseTime6,
		PreferredLifetime6: l.PreferredLifetime6,
		ValidLifetime6:     l.ValidLifetime6,
		TransactionId:      l.TransactionId,
		TransactionId6:     l.TransactionId6,
		IaId6:              l.IaId6,
		ServerId6:          l.ServerId6,
	}
}

func (l *Lease) IfaceReady() (ready4, ready6 bool) {
	l.lock.Lock()
	address := l.Address
	address6 := l.Address6
	l.lock.Unlock()

	iface, err := net.InterfaceByName(l.Iface)
	if err != nil {
		return
	}

	addrs, _ := iface.Addrs()
	for _, addr := range addrs {
		ipnet, ok := addr.(*net.IPNet)
		if ok {
			if address != nil && ipnet.IP.Equal(address.IP) {
				ready4 = true
			}

			if address6 != nil && ipnet.IP.Equal(address6.IP) {
				ready6 = true
			}
		}
	}

	return
}

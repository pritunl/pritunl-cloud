package dhcpc

import (
	"net"
	"time"

	"github.com/insomniacslk/dhcp/dhcpv6"
)

type Lease struct {
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
}

func (l *Lease) IfaceReady() (ready4, ready6 bool) {
	iface, err := net.InterfaceByName(l.Iface)
	if err != nil {
		return
	}

	addrs, _ := iface.Addrs()
	for _, addr := range addrs {
		ipnet, ok := addr.(*net.IPNet)
		if ok {
			if l.Address != nil && ipnet.IP.Equal(l.Address.IP) {
				ready4 = true
			}

			if l.Address6 != nil && ipnet.IP.Equal(l.Address6.IP) {
				ready6 = true
			}
		}
	}

	return
}

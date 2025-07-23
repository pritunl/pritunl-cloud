package dhcpc

import (
	"net"
	"time"
)

type Lease struct {
	Iface         string        `json:"iface"`
	Address       *net.IPNet    `json:"address"`
	Gateway       net.IP        `json:"gateway"`
	Address6      *net.IPNet    `json:"address6"`
	ServerAddress net.IP        `json:"server"`
	LeaseTime     time.Duration `json:"ttl"`
	TransactionId string        `json:"-"`
}

func (l *Lease) IfaceReady() (ready bool) {
	iface, err := net.InterfaceByName(l.Iface)
	if err != nil {
		return
	}

	addrs, _ := iface.Addrs()
	for _, addr := range addrs {
		ipnet, ok := addr.(*net.IPNet)
		if ok {
			if ipnet.IP.Equal(l.Address.IP) {
				ready = true
				break
			}
		}
	}

	return
}

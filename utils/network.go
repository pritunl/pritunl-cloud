package utils

import (
	"net"
)

func IncIpAddress(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func CopyIpAddress(src net.IP) net.IP {
	dst := make(net.IP, len(src))
	copy(dst, src)
	return dst
}

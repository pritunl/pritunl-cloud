package utils

import (
	"encoding/binary"
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

func IpAddress2Int(ip net.IP) int64 {
	if len(ip) == 16 {
		return int64(binary.BigEndian.Uint32(ip[12:16]))
	}
	return int64(binary.BigEndian.Uint32(ip))
}

func Int2IpAddress(n int64) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, uint32(n))
	return ip
}

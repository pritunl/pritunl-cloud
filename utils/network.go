package utils

import (
	"encoding/binary"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"io/ioutil"
	"math/big"
	"net"
	"os"
)

func IncIpAddress(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func DecIpAddress(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]--
		if ip[j] < 255 {
			break
		}
	}
}

func CopyIpAddress(src net.IP) net.IP {
	dst := make(net.IP, len(src))
	copy(dst, src)
	return dst
}

func IpAddress2BigInt(ip net.IP) (n *big.Int, bits int) {
	n = &big.Int{}
	n.SetBytes([]byte(ip))
	if len(ip) == net.IPv4len {
		bits = 32
	} else {
		bits = 128
	}
	return
}

func BigInt2IpAddress(n *big.Int, bits int) net.IP {
	byt := n.Bytes()
	ip := make([]byte, bits/8)
	for i := 1; i <= len(byt); i++ {
		ip[len(ip)-i] = byt[len(byt)-i]
	}
	return net.IP(ip)
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

func GetLastIpAddress(network *net.IPNet) net.IP {
	prefixLen, bits := network.Mask.Size()
	if prefixLen == bits {
		return CopyIpAddress(network.IP)
	}
	start, bits := IpAddress2BigInt(network.IP)
	n := uint(bits) - uint(prefixLen)
	end := big.NewInt(1)
	end.Lsh(end, n)
	end.Sub(end, big.NewInt(1))
	end.Or(end, start)
	return BigInt2IpAddress(end, bits)
}

func ParseIpMask(mask string) net.IPMask {
	maskIp := net.ParseIP(mask)
	if maskIp == nil {
		return nil
	}
	return net.IPv4Mask(maskIp[12], maskIp[13], maskIp[14], maskIp[15])
}

func GetNamespaces() (namespaces []string, err error) {
	items, err := ioutil.ReadDir("/var/run/netns")
	if err != nil {
		if os.IsNotExist(os.ErrNotExist) {
			namespaces = []string{}
			err = nil
		} else {
			err = &errortypes.ReadError{
				errors.Wrap(err, "utils: Failed to read network namespaces"),
			}
		}
		return
	}

	namespaces = []string{}
	for _, item := range items {
		namespaces = append(namespaces, item.Name())
	}

	return
}

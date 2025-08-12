package utils

import (
	"encoding/binary"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"path"
	"strings"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

var (
	private10 = net.IPNet{
		IP:   net.IPv4(10, 0, 0, 0),
		Mask: net.CIDRMask(8, 32),
	}
	private100 = net.IPNet{
		IP:   net.IPv4(100, 64, 0, 0),
		Mask: net.CIDRMask(10, 32),
	}
	private172 = net.IPNet{
		IP:   net.IPv4(172, 16, 0, 0),
		Mask: net.CIDRMask(12, 32),
	}
	private192 = net.IPNet{
		IP:   net.IPv4(192, 168, 0, 0),
		Mask: net.CIDRMask(16, 32),
	}
	private198 = net.IPNet{
		IP:   net.IPv4(198, 18, 0, 0),
		Mask: net.CIDRMask(15, 32),
	}
	reserved6 = net.IPNet{
		IP:   net.IPv4(6, 0, 0, 0),
		Mask: net.CIDRMask(8, 32),
	}
	reserved11 = net.IPNet{
		IP:   net.IPv4(11, 0, 0, 0),
		Mask: net.CIDRMask(8, 32),
	}
	reserved21 = net.IPNet{
		IP:   net.IPv4(21, 0, 0, 0),
		Mask: net.CIDRMask(8, 32),
	}
	reserved25 = net.IPNet{
		IP:   net.IPv4(25, 0, 0, 0),
		Mask: net.CIDRMask(8, 32),
	}
	reserved26 = net.IPNet{
		IP:   net.IPv4(26, 0, 0, 0),
		Mask: net.CIDRMask(8, 32),
	}
	reserved53 = net.IPNet{
		IP:   net.IPv4(53, 0, 0, 0),
		Mask: net.CIDRMask(8, 32),
	}
	reserved57 = net.IPNet{
		IP:   net.IPv4(57, 0, 0, 0),
		Mask: net.CIDRMask(8, 32),
	}

	loopback127 = net.IPNet{
		IP:   net.IPv4(127, 0, 0, 0),
		Mask: net.CIDRMask(8, 32),
	}
	linkLocal169 = net.IPNet{
		IP:   net.IPv4(169, 254, 0, 0),
		Mask: net.CIDRMask(16, 32),
	}
	multicast224 = net.IPNet{
		IP:   net.IPv4(224, 0, 0, 0),
		Mask: net.CIDRMask(4, 32),
	}
	broadcast255 = net.IPNet{
		IP:   net.IPv4(255, 255, 255, 255),
		Mask: net.CIDRMask(32, 32),
	}
	zeroconf0 = net.IPNet{
		IP:   net.IPv4(0, 0, 0, 0),
		Mask: net.CIDRMask(8, 32),
	}
)

func IsPrivateIp(ip net.IP) bool {
	if ip == nil {
		return false
	}

	if ip.To4() == nil {
		return (ip[0] & 0xfe) == 0xfc
	}

	if private10.Contains(ip) ||
		private100.Contains(ip) ||
		private172.Contains(ip) ||
		private192.Contains(ip) ||
		private198.Contains(ip) ||
		reserved6.Contains(ip) ||
		reserved11.Contains(ip) ||
		reserved21.Contains(ip) ||
		reserved25.Contains(ip) ||
		reserved26.Contains(ip) ||
		reserved53.Contains(ip) ||
		reserved57.Contains(ip) {

		return true
	}

	return false
}

func IsPublicIp(ip net.IP) bool {
	if ip == nil {
		return false
	}

	if ip.To4() == nil {
		if (ip[0] & 0xfe) == 0xfc {
			return false
		}
		if ip[0] == 0xfe && (ip[1]&0xc0) == 0x80 {
			return false
		}
		if ip.Equal(net.IPv6loopback) {
			return false
		}
		if ip.Equal(net.IPv6unspecified) {
			return false
		}
		if ip[0] == 0xff {
			return false
		}
		return true
	}

	if IsPrivateIp(ip) ||
		loopback127.Contains(ip) ||
		linkLocal169.Contains(ip) ||
		multicast224.Contains(ip) ||
		broadcast255.Contains(ip) ||
		zeroconf0.Contains(ip) {

		return false
	}

	return true
}

type Address struct {
	Address net.IP
	Network *net.IPNet
	Ip6     bool
	Private bool
	Public  bool
}

func ParseAddress(addrStr string) (addr *Address) {
	addrStr = strings.TrimSpace(addrStr)
	if addrStr == "" {
		return
	}

	if strings.Contains(addrStr, "/") {
		ip, network, err := net.ParseCIDR(addrStr)
		if err != nil {
			return
		}

		if ip == nil {
			return
		}

		addr = &Address{
			Address: ip,
			Network: network,
			Ip6:     ip.To4() == nil,
			Private: IsPrivateIp(ip),
			Public:  IsPublicIp(ip),
		}
		return
	}

	ip := net.ParseIP(addrStr)
	if ip == nil {
		return
	}

	addr = &Address{
		Address: ip,
		Ip6:     ip.To4() == nil,
		Private: IsPrivateIp(ip),
		Public:  IsPublicIp(ip),
	}
	return
}

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
	n.SetBytes(ip)
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
	return ip
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

func Int2IpIndex(n int64) (x int64, err error) {
	if n%2 != 0 {
		err = errortypes.ParseError{
			errors.Newf("utils: Odd network int divide %d", n),
		}
		return
	}

	x = n / 2
	return
}

func GetFirstIpIndex(network *net.IPNet) (n int64, err error) {
	startIp := CopyIpAddress(network.IP)
	startInt := IpAddress2Int(startIp)

	startIndex, err := Int2IpIndex(startInt)
	if err != nil {
		return
	}

	n = startIndex + 1
	return
}

func GetLastIpIndex(network *net.IPNet) (n int64, err error) {
	endIp := GetLastIpAddress(network)
	endInt := IpAddress2Int(endIp) - 1

	endIndex, err := Int2IpIndex(endInt)
	if err != nil {
		return
	}

	n = endIndex - 1
	return
}

func IpIndex2Ip(index int64) (x, y net.IP) {
	x = Int2IpAddress(index * 2)
	y = CopyIpAddress(x)
	IncIpAddress(y)
	return
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

func NetworkContains(x, y *net.IPNet) bool {
	return x.Contains(y.IP) && x.Contains(GetLastIpAddress(y))
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

func GetInterfaces() (ifaces []string, err error) {
	ifaces, _, err = GetInterfacesSet()
	if err != nil {
		return
	}

	return
}

func GetInterfacesSet() (ifaces []string, ifacesSet set.Set, err error) {
	items, err := ioutil.ReadDir("/sys/class/net")
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "utils: Failed to read network interfaces"),
		}
		return
	}

	ifaces = []string{}
	ifacesSet = set.NewSet()
	for _, item := range items {
		name := item.Name()

		if name == "" {
			continue
		}

		ifaces = append(ifaces, name)
		ifacesSet.Add(name)
	}

	exists, err := ExistsDir("/etc/sysconfig/network-scripts")
	if err != nil {
		return
	}

	if exists {
		items, err = ioutil.ReadDir("/etc/sysconfig/network-scripts")
		if err != nil {
			err = &errortypes.ReadError{
				errors.Wrap(err, "utils: Failed to read network scripts"),
			}
			return
		}

		for _, item := range items {
			name := item.Name()

			if !strings.HasPrefix(name, "ifcfg-") ||
				!strings.Contains(name, ":") {

				continue
			}

			name = name[6:]
			names := strings.Split(name, ":")
			if len(names) != 2 {
				continue
			}

			if name == "" {
				continue
			}

			if ifacesSet.Contains(names[0]) && !ifacesSet.Contains(name) {
				ifaces = append(ifaces, name)
				ifacesSet.Add(name)
			}
		}
	}

	return
}

func GetInterfaceUpper(iface string) (upper string, err error) {
	iface = strings.Split(iface, ":")[0]

	items, err := ioutil.ReadDir("/sys/class/net/" + iface)
	if err != nil {
		if os.IsNotExist(os.ErrNotExist) {
			err = nil
		} else {
			err = &errortypes.ReadError{
				errors.Wrap(err, "utils: Failed to read network interface"),
			}
		}
		return
	}

	for _, item := range items {
		name := item.Name()
		if strings.HasPrefix(name, "upper_") {
			upper = name[6:]
		}
	}

	return
}

func IsInterfaceBridge(iface string) (bridge bool, err error) {
	bridge, err = ExistsDir(
		path.Join("/", "sys", "class", "net", iface, "bridge"))
	if err != nil {
		return
	}

	return
}

func FilterIp(input string) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return ""
	}
	ip := net.ParseIP(input)
	if ip == nil {
		return ""
	}
	return ip.String()
}

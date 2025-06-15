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

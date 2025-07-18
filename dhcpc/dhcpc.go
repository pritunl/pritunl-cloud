package dhcpc

import (
	"flag"
	"net"
	"os"
	"strconv"

	"github.com/pritunl/tools/logger"
)

type Dhcpc struct {
	ImdsAddress string
	ImdsPort    int
	ImdsSecret  string
	DhcpIface   string
	DhcpIface6  string
	DhcpIp      *net.IPNet
	DhcpIp6     *net.IPNet
}

func Main() (err error) {
	imdsAddress := os.Getenv("IMDS_ADDRESS")
	imdsPort := os.Getenv("IMDS_PORT")
	imdsSecret := os.Getenv("IMDS_SECRET")
	dhcpIface := os.Getenv("DHCP_IFACE")
	dhcpIface6 := os.Getenv("DHCP_IFACE6")
	dhcpIp := os.Getenv("DHCP_IP")
	dhcpIp6 := os.Getenv("DHCP_IP6")
	os.Unsetenv("IMDS_ADDRESS")
	os.Unsetenv("IMDS_PORT")
	os.Unsetenv("IMDS_SECRET")
	os.Unsetenv("DHCP_IFACE")
	os.Unsetenv("DHCP_IFACE6")
	os.Unsetenv("DHCP_IP")
	os.Unsetenv("DHCP_IP6")

	logger.Init(
		logger.SetTimeFormat(""),
	)

	ip4 := false
	flag.BoolVar(&ip4, "ip4", false, "Enable IPv4")

	ip6 := false
	flag.BoolVar(&ip6, "ip6", false, "Enable IPv6")

	flag.Parse()

	imdsPortInt, _ := strconv.Atoi(imdsPort)

	client := &Dhcpc{
		ImdsAddress: imdsAddress,
		ImdsPort:    imdsPortInt,
		ImdsSecret:  imdsSecret,
		DhcpIface:   dhcpIface,
		DhcpIface6:  dhcpIface6,
	}

	if dhcpIp != "" {
		ip, ipnet, _ := net.ParseCIDR(dhcpIp)
		ipnet.IP = ip
		client.DhcpIp = ipnet
	}

	if dhcpIp6 != "" {
		ip, ipnet, _ := net.ParseCIDR(dhcpIp6)
		ipnet.IP = ip
		client.DhcpIp6 = ipnet
	}

	return
}

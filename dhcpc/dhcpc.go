package dhcpc

import (
	"flag"
	"net"
	"os"

	"github.com/pritunl/tools/logger"
)

func Main() (err error) {
	ImdsAddress = os.Getenv("IMDS_ADDRESS")
	ImdsPort = os.Getenv("IMDS_PORT")
	DhcpSecret = os.Getenv("DHCP_SECRET")
	DhcpIface = os.Getenv("DHCP_IFACE")
	DhcpIface6 = os.Getenv("DHCP_IFACE6")
	dhcpIp := os.Getenv("DHCP_IP")
	dhcpIp6 := os.Getenv("DHCP_IP6")
	os.Unsetenv("IMDS_ADDRESS")
	os.Unsetenv("IMDS_PORT")
	os.Unsetenv("DHCP_SECRET")
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

	lease := &Lease{}

	if dhcpIp != "" || dhcpIp6 != "" {
		ip, ipnet, _ := net.ParseCIDR(dhcpIp)
		ipnet.IP = ip
		lease.Address = ipnet
	}

	_, err = lease.Exchange()
	if err != nil {
		return
	}

	return
}

package dhcpc

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/pritunl/tools/logger"
)

func Main() (err error) {
	ImdsAddress = os.Getenv("IMDS_ADDRESS")
	ImdsPort = os.Getenv("IMDS_PORT")
	DhcpSecret = os.Getenv("DHCP_SECRET")
	DhcpIface = os.Getenv("DHCP_IFACE")
	DhcpIface6 = os.Getenv("DHCP_IFACE6")
	os.Unsetenv("IMDS_ADDRESS")
	os.Unsetenv("IMDS_PORT")
	os.Unsetenv("DHCP_SECRET")
	os.Unsetenv("DHCP_IFACE")
	os.Unsetenv("DHCP_IFACE6")

	logger.Init(
		logger.SetTimeFormat(""),
	)

	ip4 := false
	flag.BoolVar(&ip4, "ip4", false, "Enable IPv4")

	ip6 := false
	flag.BoolVar(&ip6, "ip6", false, "Enable IPv6")

	iface := ""
	flag.StringVar(&iface, "iface", "", "Bind interface")

	flag.Parse()
}

package dhcpc

import (
	"flag"
	"os"
)

func Main() {
	DhcpSecret = os.Getenv("DHCP_SECRET")
	os.Unsetenv("DHCP_SECRET")

	ip4 := false
	flag.BoolVar(&ip4, "ip4", false, "Enable IPv4")

	ip6 := false
	flag.BoolVar(&ip6, "ip6", false, "Enable IPv6")

	iface := ""
	flag.StringVar(&iface, "iface", "", "Bind interface")

	flag.Parse()
}

package server

import (
	"flag"
	"os"
	"strings"

	"github.com/pritunl/pritunl-cloud/imds/server/constants"
	"github.com/pritunl/pritunl-cloud/imds/server/router"
	"github.com/pritunl/pritunl-cloud/imds/server/state"
)

func Main() (err error) {
	constants.ClientSecret = os.Getenv("CLIENT_SECRET")
	constants.DhcpSecret = os.Getenv("DHCP_SECRET")
	constants.HostSecret = os.Getenv("HOST_SECRET")
	os.Unsetenv("CLIENT_SECRET")
	os.Unsetenv("DHCP_SECRET")
	os.Unsetenv("HOST_SECRET")

	host := ""
	flag.StringVar(&host, "host", "127.0.0.1", "Server bind address")

	port := 0
	flag.IntVar(&port, "port", 80, "Server bind port")

	client := ""
	flag.StringVar(&client, "client", "127.0.0.1", "Client address")

	sockPath := ""
	flag.StringVar(&sockPath, "sock", "", "Socket path")

	flag.Parse()

	constants.Host = strings.Split(host, "/")[0]
	constants.Port = port
	constants.Sock = sockPath
	constants.Client = client

	routr := &router.Router{}
	routr.Init()

	err = state.Init()
	if err != nil {
		return
	}

	err = routr.Run()
	if err != nil {
		return
	}

	return
}

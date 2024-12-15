package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/pritunl/pritunl-cloud/imds/server/constants"
	"github.com/pritunl/pritunl-cloud/imds/server/router"
	"github.com/pritunl/pritunl-cloud/imds/server/state"
)

const help = `
Usage: pritunl-cloud-imds COMMAND

Commands:
  version      Show version
  start        Start IMDS server
`

func main() {
	flag.Usage = func() {
		fmt.Printf(help)
	}

	host := ""
	flag.StringVar(&host, "host", "127.0.0.1", "Server bind address")

	port := 0
	flag.IntVar(&port, "port", 80, "Server bind port")

	client := ""
	flag.StringVar(&host, "client", "127.0.0.1", "Client address")

	sockPath := ""
	flag.StringVar(&sockPath, "sock", "", "Socket path")

	flag.Parse()

	switch flag.Arg(0) {
	case "start":
		constants.Host = strings.Split(host, "/")[0]
		constants.Port = port
		constants.Sock = sockPath
		constants.Client = client
		constants.Secret = os.Getenv("SECRET")
		config.Path = confPath
		state.Path = statePath
		config.Path = confPath

		routr := &router.Router{}
		routr.Init()

		err := config.Init()
		if err != nil {
			panic(err)
		}

		err = state.Init()
		if err != nil {
			panic(err)
		}

		err = routr.Run()
		if err != nil {
			panic(err)
		}
		break
	case "version":
		fmt.Printf("pritunl-cloud-imds v%s\n", constants.Version)
		break
	}
}

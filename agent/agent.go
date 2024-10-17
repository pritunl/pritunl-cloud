package main

import (
	"flag"
	"fmt"

	"github.com/pritunl/pritunl-cloud/agent/constants"
	"github.com/pritunl/pritunl-cloud/agent/engine"
	"github.com/pritunl/pritunl-cloud/agent/imds"
)

const help = `
Usage: pci COMMAND

Commands:
  get          Get value from IMDS
  version      Show version
`

func main() {
	flag.Usage = func() {
		fmt.Printf(help)
	}

	flag.Parse()

	switch flag.Arg(0) {
	case "get":
		ids := &imds.Imds{}

		err := ids.Init()
		if err != nil {
			panic(err)
		}

		val, err := ids.Get(flag.Arg(1))
		if err != nil {
			panic(err)
		}

		fmt.Print(val)

		break
	case "engine":
		eng := &engine.Engine{}

		err := eng.Init(flag.Arg(1))
		if err != nil {
			panic(err)
		}
		break
	case "version":
		fmt.Printf("pci v%s\n", constants.Version)
		break
	default:
		fmt.Printf(help)
	}
}

package main

import (
	"flag"
	"fmt"

	"github.com/pritunl/pritunl-cloud/agent/constants"
	"github.com/pritunl/pritunl-cloud/agent/engine"
)

const help = `
Usage: pci COMMAND

Commands:
  version      Show version
`

func main() {
	switch flag.Arg(0) {
	case "engine":
		eng := &engine.Engine{}

		err := eng.Init()
		if err != nil {
			panic(err)
		}
		break
	case "version":
		fmt.Printf("pci v%s\n", constants.Version)
		break
	}
}

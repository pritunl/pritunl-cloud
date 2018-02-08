package main

import (
	"flag"
	"fmt"
	"github.com/pritunl/pritunl-cloud/constants"
	"time"
)

const help = `
Usage: pritunl-cloud COMMAND

Commands:
  version     Show version
  start       Start node
`

func main() {
	defer time.Sleep(500 * time.Millisecond)

	flag.Parse()

	switch flag.Arg(0) {
	case "start":
		return
	case "version":
		fmt.Printf("pritunl-cloud v%s\n", constants.Version)
		return
	}

	fmt.Println(help)
}

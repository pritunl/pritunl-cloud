package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/pritunl/pritunl-cloud/agent/constants"
	"github.com/pritunl/pritunl-cloud/agent/engine"
	"github.com/pritunl/pritunl-cloud/agent/imds"
	"github.com/pritunl/pritunl-cloud/agent/logging"
	"github.com/pritunl/tools/logger"
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
		ids := &imds.Imds{}
		log := &logging.Redirect{}

		err := log.Open()
		if err != nil {
			panic(err)
		}
		defer log.Close()

		err = ids.Init()
		if err != nil {
			logger.WithFields(logger.Fields{
				"error": err,
			}).Error("agent: Failed to init imds")
			panic(err)
		}

		for i := 0; i < 150; i++ {
			time.Sleep(200 * time.Millisecond)

			err = ids.Sync()
			if err != nil {
				continue
			}

			break
		}
		if err != nil {
			logger.WithFields(logger.Fields{
				"error": err,
			}).Error("agent: Failed to sync imds initial")
			panic(err)
		}

		err = eng.Init(flag.Arg(1))
		if err != nil {
			logger.WithFields(logger.Fields{
				"error": err,
			}).Error("agent: Failed to init engine")
			panic(err)
		}

		err = ids.Run()
		if err != nil {
			logger.WithFields(logger.Fields{
				"error": err,
			}).Error("agent: Failed to run")
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

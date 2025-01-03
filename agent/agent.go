package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/agent/constants"
	"github.com/pritunl/pritunl-cloud/agent/imds"
	"github.com/pritunl/pritunl-cloud/engine"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/imds/types"
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

	logger.Init(
		logger.SetTimeFormat(""),
	)

	switch flag.Arg(0) {
	case "get":
		ids := &imds.Imds{}

		err := ids.Init(nil)
		if err != nil {
			panic(err)
		}
		defer ids.Close()

		val, err := ids.Get(flag.Arg(1))
		if err != nil {
			panic(err)
		}

		fmt.Print(val)

		break
	case "engine":
		eng := &engine.Engine{}
		ids := &imds.Imds{}

		err := ids.Init(eng)
		if err != nil {
			logger.WithFields(logger.Fields{
				"error": err,
			}).Error("agent: Failed to init imds")
			panic(err)
		}
		defer ids.Close()

		ready := false
		for i := 0; i < 900; i++ {
			time.Sleep(200 * time.Millisecond)

			ready, err = ids.Sync()
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
		} else if !ready {
			err = &errortypes.RequestError{
				errors.New("agent: Initial config timeout"),
			}
			logger.WithFields(logger.Fields{
				"error": err,
			}).Error("agent: Timeout waiting for imds initial config")
			panic(err)
		}

		err = eng.Init()
		if err != nil {
			logger.WithFields(logger.Fields{
				"error": err,
			}).Error("agent: Failed to init engine")
			panic(err)
		}

		image := false
		phase := engine.Reboot
		switch flag.Arg(1) {
		case engine.Image:
			image = true
			phase = engine.Initial
			break
		case engine.Initial:
			phase = engine.Initial
			break
		}

		ids.RunSync(image)

		err = eng.Run(phase)
		if err != nil {
			return
		}

		if !image {
			ids.SetInitialized()

			err = ids.SyncStatus(types.Running)
			if err != nil {
				logger.WithFields(logger.Fields{
					"error": err,
				}).Error("agent: Failed to sync status")
				panic(err)
			}

			err = ids.Wait()
			if err != nil {
				logger.WithFields(logger.Fields{
					"error": err,
				}).Error("agent: Failed to run")
				panic(err)
			}
		}

		time.Sleep(500 * time.Millisecond)

		_, err = ids.Sync()
		if err != nil {
			logger.WithFields(logger.Fields{
				"error": err,
			}).Error("agent: Failed to sync")
			panic(err)
		}

		break
	case "image":
		ids := &imds.Imds{}

		err := ids.Init(nil)
		if err != nil {
			panic(err)
		}

		err = ids.SyncStatus(types.Imaged)
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

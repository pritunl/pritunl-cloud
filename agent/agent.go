package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/agent/constants"
	"github.com/pritunl/pritunl-cloud/agent/imds"
	"github.com/pritunl/pritunl-cloud/agent/utils"
	"github.com/pritunl/pritunl-cloud/engine"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/imds/types"
	pritunl_utils "github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/tools/logger"
)

const help = `
Usage: pci COMMAND

Commands:
  get          Get value from IMDS
  image        Sanitize host files and initiate shutdown for imaging
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

	logger.AddHandler(func(record *logger.Record) {
		fmt.Print(record.String())
	})

	switch flag.Arg(0) {
	case "get":
		ids := &imds.Imds{}

		err := ids.Init(nil)
		if err != nil {
			logger.WithFields(logger.Fields{
				"error": err,
			}).Error("agent: Initialize failed")
			utils.DelayExit(1, 1*time.Second)
			return
		}
		defer ids.Close()

		val, err := ids.Get(flag.Arg(1))
		if err != nil {
			logger.WithFields(logger.Fields{
				"error": err,
			}).Error("agent: Get imds failed")
			utils.DelayExit(1, 1*time.Second)
			return
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
			utils.DelayExit(1, 1*time.Second)
			return
		}
		defer ids.Close()

		err = ids.OpenLog()
		if err != nil {
			return
		}

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
			utils.DelayExit(1, 1*time.Second)
			return
		} else if !ready {
			err = &errortypes.RequestError{
				errors.New("agent: Initial config timeout"),
			}
			logger.WithFields(logger.Fields{
				"error": err,
			}).Error("agent: Timeout waiting for imds initial config")
			utils.DelayExit(1, 1*time.Second)
			return
		}

		err = eng.Init()
		if err != nil {
			logger.WithFields(logger.Fields{
				"error": err,
			}).Error("agent: Failed to init engine")
			utils.DelayExit(1, 1*time.Second)
			return
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
			logger.WithFields(logger.Fields{
				"error": err,
			}).Error("agent: Engine run failed")
			utils.DelayExit(1, 1*time.Second)
			return
		}

		if !image {
			ids.SetInitialized()

			err = ids.SyncStatus(types.Running)
			if err != nil {
				logger.WithFields(logger.Fields{
					"error": err,
				}).Error("agent: Failed to sync status")
				utils.DelayExit(1, 1*time.Second)
				return
			}

			err = ids.Wait()
			if err != nil {
				logger.WithFields(logger.Fields{
					"error": err,
				}).Error("agent: Failed to run")
				utils.DelayExit(1, 1*time.Second)
				return
			}
		}

		time.Sleep(500 * time.Millisecond)

		_, err = ids.Sync()
		if err != nil {
			logger.WithFields(logger.Fields{
				"error": err,
			}).Error("agent: Failed to sync")
			utils.DelayExit(1, 1*time.Second)
			return
		}

		break
	case "image":
		ids := &imds.Imds{}

		err := ids.Init(nil)
		if err != nil {
			logger.WithFields(logger.Fields{
				"error": err,
			}).Error("agent: Iniatilize failed")
			utils.DelayExit(1, 1*time.Second)
			return
		}

		err = utils.Sanitize()
		if err != nil {
			logger.WithFields(logger.Fields{
				"error": err,
			}).Error("agent: Sanitize failed")
			utils.DelayExit(1, 1*time.Second)
			return
		}

		err = ids.SyncStatus(types.Imaged)
		if err != nil {
			logger.WithFields(logger.Fields{
				"error": err,
			}).Error("agent: Sync status failed")
			utils.DelayExit(1, 1*time.Second)
			return
		}

		err = utils.SanitizeImds()
		if err != nil {
			logger.WithFields(logger.Fields{
				"error": err,
			}).Error("agent: Sanitize imds failed")
			utils.DelayExit(1, 1*time.Second)
			return
		}

		break
	case "status":
		mem, err := pritunl_utils.GetMemInfo()
		if err != nil {
			logger.WithFields(logger.Fields{
				"error": err,
			}).Error("imds: Failed to get memory")
			utils.DelayExit(1, 1*time.Second)
			return
		}

		fmt.Println("Memory Information:")
		fmt.Printf("  Total Memory:           %d KB\n", mem.Total)
		fmt.Printf("  Free Memory:            %d KB\n", mem.Free)
		fmt.Printf("  Available Memory:       %d KB\n", mem.Available)
		fmt.Printf("  Buffers:                %d KB\n", mem.Buffers)
		fmt.Printf("  Cached:                 %d KB\n", mem.Cached)
		fmt.Printf("  Used Memory:            %d KB\n", mem.Used)
		fmt.Printf("  Used Percentage:        %.2f%%\n", mem.UsedPercent)
		fmt.Printf("  Dirty:                  %d KB\n", mem.Dirty)
		fmt.Println("\nSwap Information:")
		fmt.Printf("  Swap Total:             %d KB\n", mem.SwapTotal)
		fmt.Printf("  Swap Free:              %d KB\n", mem.SwapFree)
		fmt.Printf("  Swap Used:              %d KB\n", mem.SwapUsed)
		fmt.Printf("  Swap Used Percentage:   %.2f%%\n", mem.SwapUsedPercent)
		fmt.Println("\nHugePages Information:")
		fmt.Printf("  HugePages Total:        %d\n", mem.HugePagesTotal)
		fmt.Printf("  HugePages Free:         %d\n", mem.HugePagesFree)
		fmt.Printf("  HugePages Reserved:     %d\n", mem.HugePagesReserved)
		fmt.Printf("  HugePages Used:         %d\n", mem.HugePagesUsed)
		fmt.Printf("  HugePages Used Percent: %.2f%%\n", mem.HugePagesUsedPercent)
		fmt.Printf("  HugePage Size:          %d KB\n", mem.HugePageSize)

		load, err := pritunl_utils.LoadAverage()
		if err != nil {
			logger.WithFields(logger.Fields{
				"error": err,
			}).Error("imds: Failed to get load")
			utils.DelayExit(1, 1*time.Second)
			return
		}

		fmt.Println("\nLoad Average Information:")
		fmt.Printf("  CPU Units:              %d\n", load.CpuUnits)
		fmt.Printf("  Load Average (1 min):   %.2f%%\n", load.Load1)
		fmt.Printf("  Load Average (5 min):   %.2f%%\n", load.Load5)
		fmt.Printf("  Load Average (15 min):  %.2f%%\n", load.Load15)
		break
	case "version":
		fmt.Printf("pci v%s\n", constants.Version)
		break
	default:
		fmt.Printf(help)
	}

	return
}

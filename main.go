package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pritunl/pritunl-cloud/cmd"
	"github.com/pritunl/pritunl-cloud/constants"
	"github.com/pritunl/pritunl-cloud/logger"
	"github.com/pritunl/pritunl-cloud/requires"
)

const help = `
Usage: pritunl-cloud COMMAND

Commands:
  version           Show version
  mongo             Set MongoDB URI
  set               Set a setting
  unset             Unset a setting
  start             Start node
  clear-logs        Clear logs
  reset-node-web    Reset node web server settings
  optimize          Optimize system configuration
  default-password  Get default administrator password
  reset-password    Reset administrator password
  disable-policies  Disable all policies
  disable-firewall  Disable firewall on this node
  start-instance    Start instance by name
  stop-instance     Stop instance by name
  mtu-check         Check and show instance MTUs
  backup            Backup local data
`

func Init() {
	logger.Init()
	requires.Init(nil)
}

func InitLimited() {
	logger.Init()
	requires.Init([]string{"ahandlers", "uhandlers"})
}

func main() {
	defer time.Sleep(500 * time.Millisecond)

	command := ""
	for _, arg := range os.Args[1:] {
		if !strings.HasPrefix(arg, "-") {
			command = arg
			break
		}
	}

	switch command {
	case "start":
		flag.Parse()

		for _, arg := range flag.Args() {
			switch arg {
			case "--debug":
				constants.Production = false
				break
			case "--debug-web":
				constants.DebugWeb = true
				break
			case "--fast-exit":
				constants.FastExit = true
				break
			}
		}

		Init()
		err := cmd.Node()
		if err != nil {
			panic(err)
		}
		return
	case "version":
		fmt.Printf("pritunl-cloud v%s\n", constants.Version)
		return
	case "mongo":
		flag.Parse()
		logger.Init()
		err := cmd.Mongo()
		if err != nil {
			panic(err)
		}
		return
	case "optimize":
		logger.Init()
		err := cmd.Optimize()
		if err != nil {
			panic(err)
		}
		return
	case "reset-node-web":
		InitLimited()
		err := cmd.ResetNodeWeb()
		if err != nil {
			panic(err)
		}
		return
	case "default-password":
		InitLimited()
		err := cmd.DefaultPassword()
		if err != nil {
			panic(err)
		}
		return
	case "reset-password":
		InitLimited()
		err := cmd.ResetPassword()
		if err != nil {
			panic(err)
		}
		return
	case "disable-policies":
		InitLimited()
		err := cmd.DisablePolicies()
		if err != nil {
			panic(err)
		}
		return
	case "disable-firewall":
		InitLimited()
		err := cmd.DisableFirewall()
		if err != nil {
			panic(err)
		}
		return
	case "mtu-check":
		InitLimited()
		err := cmd.MtuCheck()
		if err != nil {
			panic(err)
		}
		return
	case "set":
		flag.Parse()
		InitLimited()
		err := cmd.SettingsSet()
		if err != nil {
			panic(err)
		}
		return
	case "unset":
		flag.Parse()
		InitLimited()
		err := cmd.SettingsUnset()
		if err != nil {
			panic(err)
		}
		return
	case "clear-logs":
		InitLimited()
		err := cmd.ClearLogs()
		if err != nil {
			panic(err)
		}
		return
	case "backup":
		flag.Parse()
		InitLimited()
		err := cmd.Backup()
		if err != nil {
			panic(err)
		}
		return
	case "imds-server":
		err := cmd.ImdsServer()
		if err != nil {
			panic(err)
		}
		return
	case "dhcp4-server":
		err := cmd.Dhcp4Server()
		if err != nil {
			panic(err)
		}
		return
	case "dhcp6-server":
		err := cmd.Dhcp6Server()
		if err != nil {
			panic(err)
		}
		return
	case "dhcp-client":
		err := cmd.DhcpClient()
		if err != nil {
			panic(err)
		}
		return
	case "ndp-server":
		err := cmd.NdpServer()
		if err != nil {
			panic(err)
		}
		return
	case "start-instance":
		flag.Parse()
		InitLimited()
		err := cmd.StartInstance(flag.Args()[1])
		if err != nil {
			panic(err)
		}
		return
	case "stop-instance":
		flag.Parse()
		InitLimited()
		err := cmd.StopInstance(flag.Args()[1])
		if err != nil {
			panic(err)
		}
		return
	default:
		fmt.Printf(help)
	}
}

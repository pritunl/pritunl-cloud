package main

import (
	"flag"
	"fmt"

	"github.com/pritunl/pritunl-cloud/builder/prompt"

	"github.com/pritunl/pritunl-cloud/builder/cloud"
	"github.com/pritunl/pritunl-cloud/builder/mongo"
	"github.com/pritunl/pritunl-cloud/builder/start"
	"github.com/pritunl/pritunl-cloud/builder/sysctl"
	"github.com/pritunl/pritunl-cloud/builder/systemctl"
	"github.com/pritunl/pritunl-cloud/colorize"
	"github.com/pritunl/pritunl-cloud/logger"
)

const (
	art = `
                     /$$   /$$                         /$$
                    |__/  | $$                        | $$
  /$$$$$$   /$$$$$$  /$$ /$$$$$$   /$$   /$$ /$$$$$$$ | $$
 /$$__  $$ /$$__  $$| $$|_  $$_/  | $$  | $$| $$__  $$| $$
| $$  \ $$| $$  \__/| $$  | $$    | $$  | $$| $$  \ $$| $$
| $$  | $$| $$      | $$  | $$ /$$| $$  | $$| $$  | $$| $$
| $$$$$$$/| $$      | $$  |  $$$$/|  $$$$$$/| $$  | $$| $$
| $$____/ |__/      |__/   \___/   \______/ |__/  |__/|__/
| $$                                                      
| $$             /$$                           /$$        
|__/            | $$                          | $$        
        /$$$$$$$| $$  /$$$$$$  /$$   /$$  /$$$$$$$        
       /$$_____/| $$ /$$__  $$| $$  | $$ /$$__  $$        
      | $$      | $$| $$  \ $$| $$  | $$| $$  | $$        
      | $$      | $$| $$  | $$| $$  | $$| $$  | $$        
      |  $$$$$$$| $$|  $$$$$$/|  $$$$$$/|  $$$$$$$        
       \_______/|__/ \______/  \______/  \_______/        
`
)

const help = `
Usage: pritunl-builder OPTIONS

Options:
  --assume-yes  Assume yes to prompts
  --no-start    Do not start Pritunl Cloud service
  --unstable    Use unstable repository
`

func main() {
	logger.InitStdout()

	intro := colorize.ColorString(art, colorize.BlueBold, colorize.None)
	assumeYes := flag.Bool("assume-yes", false, "Assume yes to prompts")
	noStart := flag.Bool("no-start", false,
		"Do not start Pritunl Cloud service")
	unstable := flag.Bool("unstable", false,
		"Use unstable repository")

	flag.Usage = func() {
		fmt.Println(help)
	}

	flag.Parse()

	prompt.AssumeYes = *assumeYes

	fmt.Println(intro)

	err := sysctl.Sysctl()
	if err != nil {
		panic(err)
	}

	err = systemctl.Systemctl()
	if err != nil {
		panic(err)
	}

	err = mongo.Mongo()
	if err != nil {
		panic(err)
	}

	err = cloud.Cloud(unstable)
	if err != nil {
		panic(err)
	}

	err = start.Start(*noStart)
	if err != nil {
		panic(err)
	}
}

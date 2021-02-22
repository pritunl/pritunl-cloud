package main

import (
	"fmt"

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

func main() {
	logger.InitStdout()

	intro := colorize.ColorString(art, colorize.BlueBold, colorize.None)

	fmt.Println(intro)

	err := sysctl.Sysctl()
	if err != nil {
		panic(err)
	}

	err = systemctl.Systemctl()
	if err != nil {
		panic(err)
	}
}

package sync

import (
	"github.com/Sirupsen/logrus"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/utils"
	"strings"
	"time"
)

func interfaceUpdate() (err error) {
	output, err := utils.ExecCombinedOutput("", "route", "-n")
	if err != nil {
		return
	}

	defaultIface := ""
	defaultGateway := ""
	outputLines := strings.Split(output, "\n")
	for _, line := range outputLines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		if fields[0] == "0.0.0.0" {
			defaultIface = strings.TrimSpace(fields[len(fields)-1])
			defaultGateway = strings.TrimSpace(fields[1])
		}
	}

	node.DefaultInterface = defaultIface
	node.DefaultGateway = defaultGateway

	return
}

func interfaceRunner() {
	for {
		err := interfaceUpdate()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("sync: Failed to update interface")
			time.Sleep(3 * time.Second)
			continue
		}

		time.Sleep(10 * time.Second)
	}
}

func initInterface() {
	go interfaceRunner()
}

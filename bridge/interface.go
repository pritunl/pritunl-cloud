package bridge

import (
	"strings"

	"github.com/pritunl/pritunl-cloud/utils"
)

func getDefault() (iface string, err error) {
	output, err := utils.ExecCombinedOutput("", "route", "-n")
	if err != nil {
		return
	}

	outputLines := strings.Split(output, "\n")
	for _, line := range outputLines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		if fields[0] == "0.0.0.0" {
			iface = strings.TrimSpace(fields[len(fields)-1])
			_ = strings.TrimSpace(fields[1]) // gateway
		}
	}

	return
}

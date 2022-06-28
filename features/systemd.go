package features

import (
	"strconv"
	"strings"

	"github.com/pritunl/pritunl-cloud/utils"
)

func GetSystemdVersion() (ver int) {
	output, _ := utils.ExecCombinedOutputLogged(
		nil,
		"/usr/bin/systemctl", "--version",
	)

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if !strings.Contains(line, "systemd") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		n, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}

		ver = n
		break
	}

	return
}

func HasSystemdNamespace() bool {
	ver := GetSystemdVersion()
	if ver >= 243 {
		return true
	}
	return false
}

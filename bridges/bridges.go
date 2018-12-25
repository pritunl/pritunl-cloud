package bridges

import (
	"github.com/pritunl/pritunl-cloud/utils"
	"strings"
	"time"
)

var (
	bridges  = []string{}
	lastSync time.Time
)

func GetBridges() (brdgs []string, err error) {
	if time.Since(lastSync) < 60*time.Second {
		brdgs = bridges
		return
	}

	bridgesNew := []string{}

	output, err := utils.ExecOutput("", "brctl", "show")
	if err != nil {
		return
	}

	for i, line := range strings.Split(output, "\n") {
		if i == 0 || strings.HasPrefix(line, " ") ||
			strings.HasPrefix(line, "	") {

			continue
		}

		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) == 0 {
			continue
		}

		bridgesNew = append(bridgesNew, fields[0])
	}

	bridges = bridgesNew
	lastSync = time.Now()
	brdgs = bridges

	return
}

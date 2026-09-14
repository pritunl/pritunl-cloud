package psutil

import (
	"strings"
	"time"

	"github.com/pritunl/tools/commander"
)

func moduleNames() (names []string, err error) {
	resp, err := commander.Exec(&commander.Opt{
		Name:    "kldstat",
		Timeout: 10 * time.Second,
		PipeOut: true,
		PipeErr: true,
	})
	if err != nil {
		return
	}

	seen := map[string]bool{}
	names = []string{}

	for _, line := range strings.Split(string(resp.Output), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 5 || fields[0] == "Id" {
			continue
		}

		name := strings.TrimSuffix(fields[4], ".ko")
		if name == "" || name == "kernel" || seen[name] {
			continue
		}

		seen[name] = true
		names = append(names, name)
	}

	return
}

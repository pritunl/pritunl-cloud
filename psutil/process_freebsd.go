package psutil

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/pritunl/tools/commander"
)

func processNames() (names []string, err error) {
	resp, err := commander.Exec(&commander.Opt{
		Name: "ps",
		Args: []string{
			"-axo",
			"comm=",
		},
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
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "[") {
			continue
		}

		name := filepath.Base(line)
		if name == "" || name == "." || name == "/" || seen[name] {
			continue
		}

		seen[name] = true
		names = append(names, name)
	}

	return
}

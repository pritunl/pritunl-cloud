package telemetry

import (
	"bytes"
	"os"
	"time"

	"github.com/pritunl/tools/commander"
)

var (
	hasSevs = 0
)

func IsDnf() bool {
	_, err := os.Stat("/usr/bin/dnf")
	return err == nil
}

func HasSevs() bool {
	if hasSevs == 1 {
		return false
	} else if hasSevs == 2 {
		return true
	}

	resp, err := commander.Exec(&commander.Opt{
		Name: "dnf",
		Args: []string{
			"updateinfo",
			"list",
			"--help",
		},
		Timeout: 8 * time.Second,
		PipeOut: true,
		PipeErr: true,
	})
	if err != nil {
		hasSevs = 1
		return false
	}

	if bytes.Contains(resp.Output, []byte("--advisory-severities")) {
		hasSevs = 2
		return true
	}

	return false
}

package telemetry

import (
	"bytes"
	"os"
	"strings"
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

func matchAdvisory(id string) bool {
	return strings.HasPrefix(id, "RHSA-") ||
		strings.HasPrefix(id, "ALSA-") ||
		strings.HasPrefix(id, "RLSA-") ||
		strings.HasPrefix(id, "ELSA-") ||
		strings.HasPrefix(id, "FEDORA-")
}

func parseSeverity(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "critical":
		return Critical
	case "important":
		return Important
	case "moderate":
		return Moderate
	}
	return ""
}

func rankSeverity(severity string) int {
	switch severity {
	case Critical:
		return 0
	case Important:
		return 1
	case Moderate:
		return 2
	default:
		return 3
	}
}

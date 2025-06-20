package telemetry

import (
	"os"
)

func IsDnf() bool {
	_, err := os.Stat("/usr/bin/dnf")
	return err == nil
}

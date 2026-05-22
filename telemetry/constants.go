package telemetry

import (
	"regexp"
)

const (
	Moderate  = "moderate"
	Important = "important"
	Critical  = "critical"
)

var (
	idReg = regexp.MustCompile(
		`^[a-zA-Z]{1,10}-[a-zA-Z0-9]{1,12}-[a-zA-Z0-9]{1,12}$`)
)

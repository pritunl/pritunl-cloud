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
	idReg = regexp.MustCompile(`[^a-zA-Z0-9\-:]`)
)

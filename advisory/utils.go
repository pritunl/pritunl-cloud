package advisory

import (
	"strings"
)

func normalizeStatus(status string) string {
	switch status {
	case "Analyzed":
		return Analyzed
	case "Awaiting Analysis":
		return AwaitingAnalysis
	case "Rejected":
		return Rejected
	case "Undergoing Analysis":
		return Undergoing
	case "Modified":
		return Modified
	case "Deferred":
		return Deferred
	default:
		return strings.ToLower(strings.ReplaceAll(status, " ", "_"))
	}
}

func normalizeValue(val string) string {
	switch strings.ToUpper(val) {
	case "NONE":
		return None
	case "LOW":
		return Low
	case "MEDIUM":
		return Medium
	case "HIGH":
		return High
	case "CRITICAL":
		return Critical
	case "NETWORK":
		return Network
	case "ADJACENT_NETWORK", "ADJACENT":
		return Adjacent
	case "LOCAL":
		return Local
	case "PHYSICAL":
		return Physical
	case "REQUIRED":
		return Required
	case "UNCHANGED":
		return Unchanged
	case "CHANGED":
		return Changed
	default:
		return strings.ToLower(val)
	}
}

package advisory

import (
	"regexp"

	"github.com/dropbox/godropbox/container/set"
)

const (
	None     = "none"
	Low      = "low"
	Medium   = "medium"
	High     = "high"
	Critical = "critical"
	Network  = "network"
	Adjacent = "adjacent"
	Local    = "local"
	Physical = "physical"
	Required = "required"

	Unchanged = "unchanged"
	Changed   = "changed"

	Analyzed         = "analyzed"
	AwaitingAnalysis = "awaiting_analysis"
	Rejected         = "rejected"
	Undergoing       = "undergoing_analysis"
	Modified         = "modified"
	Deferred         = "deferred"

	nvdApi = "https://services.nvd.nist.gov/rest/json/cves/2.0"
)

var (
	cveIdReg = regexp.MustCompile(`^CVE-\d{4}-\d{4,}$`)

	ValidStatuses = set.NewSet(
		Analyzed,
		AwaitingAnalysis,
		Rejected,
		Undergoing,
		Modified,
		Deferred,
	)

	ValidSeverities = set.NewSet(
		None,
		Low,
		Medium,
		High,
		Critical,
	)

	ValidVectors = set.NewSet(
		Network,
		Adjacent,
		Local,
		Physical,
	)

	ValidComplexities = set.NewSet(
		Low,
		High,
	)

	ValidPrivileges = set.NewSet(
		None,
		Low,
		High,
	)

	ValidInteractions = set.NewSet(
		None,
		Required,
	)

	ValidScopes = set.NewSet(
		Unchanged,
		Changed,
	)

	ValidImpacts = set.NewSet(
		None,
		Low,
		High,
	)
)

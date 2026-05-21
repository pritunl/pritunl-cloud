package advisory

import (
	"regexp"

	"github.com/dropbox/godropbox/container/set"
)

const (
	None      = "none"
	Low       = "low"
	Medium    = "medium"
	High      = "high"
	Critical  = "critical"
	Network   = "network"
	Adjacent  = "adjacent"
	Local     = "local"
	Physical  = "physical"
	Required  = "required"
	Unchanged = "unchanged"
	Changed   = "changed"

	Analyzed         = "analyzed"
	AwaitingAnalysis = "awaiting_analysis"
	Rejected         = "rejected"
	Undergoing       = "undergoing_analysis"
	Modified         = "modified"
	Deferred         = "deferred"
	Pending          = "pending"

	Nist   = "nist"
	RedHat = "redhat"

	nvdApi    = "https://services.nvd.nist.gov/rest/json/cves/2.0"
	redhatApi = "https://access.redhat.com/hydra/rest/securitydata/cve/%s.json"
)

var (
	idReg = regexp.MustCompile(
		`^[a-zA-Z]{1,10}-[a-zA-Z0-9]{1,12}-[a-zA-Z0-9]{1,12}$`)

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

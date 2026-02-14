package advisory

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

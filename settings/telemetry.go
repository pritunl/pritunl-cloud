package settings

var Telemetry *telemetry

type telemetry struct {
	Id                string `bson:"_id"`
	CveSource         string `bson:"cve_source" default:"redhat"`
	CveSourceFallback string `bson:"cve_source_fallback" default:"nist"`
	NvdTtl            int    `bson:"nvd_ttl" default:"21600"`
	NvdFinalTtl       int    `bson:"nvd_final_ttl" default:"604800"`
	NvdApiLimit       int    `bson:"nvd_api_limit" default:"8"`
	NvdApiAuthLimit   int    `bson:"nvd_api_auth_limit" default:"1"`
	NvdApiKey         string `bson:"nvd_api_key"`
	RedhatTtl         int    `bson:"redhat_ttl" default:"21600"`
	RedhatFinalTtl    int    `bson:"redhat_final_ttl" default:"604800"`
	RedhatApiLimit    int    `bson:"redhat_api_limit" default:"1"`
	AnalysisTtl       int    `bson:"analysis_ttl" default:"1200"`
	SyncMaxAttempts   int    `bson:"sync_max_attempts" default:"8"`
	SyncRetryTtl      int    `bson:"sync_retry_ttl" default:"3600"`
	SyncMissingTtl    int    `bson:"sync_missing_ttl" default:"86400"`
	SyncFetchTimeout  int    `bson:"sync_fetch_timeout" default:"600"`
	SyncReferenceTtl  int    `bson:"sync_reference_ttl" default:"172800"`
	SyncThreads       int    `bson:"sync_threads" default:"8"`
	SyncWorkers       int    `bson:"sync_workers" default:"2"`
	DescriptionLimit  int    `bson:"description_limit" default:"10000"`
	UpdateCveLimit    int    `bson:"update_cve_limit" default:"5000"`
	VuxmlSizeLimit    int    `bson:"vuxml_size_limit" default:"67108864"`
}

func newTelemetry() interface{} {
	return &telemetry{
		Id: "telemetry",
	}
}

func updateTelemetry(data interface{}) {
	Telemetry = data.(*telemetry)
}

func init() {
	register("telemetry", newTelemetry, updateTelemetry)
}

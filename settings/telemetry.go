package settings

var Telemetry *telemetry

type telemetry struct {
	Id               string `bson:"_id"`
	CveSource        string `bson:"cve_source" default:"nist"`
	NvdTtl           int    `bson:"nvd_ttl" default:"21600"`
	NvdFinalTtl      int    `bson:"nvd_final_ttl" default:"604800"`
	NvdApiLimit      int    `bson:"nvd_api_limit" default:"8"`
	NvdApiAuthLimit  int    `bson:"nvd_api_auth_limit" default:"1"`
	NvdApiKey        string `bson:"nvd_api_key"`
	RedhatTtl        int    `bson:"redhat_ttl" default:"21600"`
	RedhatFinalTtl   int    `bson:"redhat_final_ttl" default:"604800"`
	RedhatApiLimit   int    `bson:"redhat_api_limit" default:"1"`
	DescriptionLimit int    `bson:"description_limit" default:"10000"`
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

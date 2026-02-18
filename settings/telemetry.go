package settings

var Telemetry *telemetry

type telemetry struct {
	Id              string `bson:"_id"`
	NvdTtl          int    `bson:"nvd_ttl" default:"21600"`
	NvdFinalTtl     int    `bson:"nvd_final_ttl" default:"604800"`
	NvdApiLimit     int    `bson:"nvd_api_limit" default:"8"`
	NvdApiAuthLimit int    `bson:"nvd_api_auth_limit" default:"1"`
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

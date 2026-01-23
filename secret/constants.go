package secret

import "github.com/pritunl/mongo-go-driver/v2/bson"

const (
	AWS         = "aws"
	Cloudflare  = "cloudflare"
	OracleCloud = "oracle_cloud"
	GoogleCloud = "google_cloud"
	Json        = "json"
)

var (
	Global = bson.NilObjectID
)

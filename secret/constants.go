package secret

import (
	"github.com/pritunl/mongo-go-driver/bson/primitive"
)

const (
	AWS         = "aws"
	Cloudflare  = "cloudflare"
	OracleCloud = "oracle_cloud"
	Json        = "json"
)

var (
	Global = primitive.NilObjectID
)

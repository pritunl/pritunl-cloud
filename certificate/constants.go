package certificate

import (
	"github.com/pritunl/mongo-go-driver/bson/primitive"
)

const (
	Text        = "text"
	LetsEncrypt = "lets_encrypt"

	AcmeHTTP = "acme_http"
	AcmeDNS  = "acme_dns"

	AcmeAWS         = "acme_aws"
	AcmeCloudflare  = "acme_cloudflare"
	AcmeOracleCloud = "acme_oracle_cloud"
)

var (
	Global = primitive.NilObjectID
)

package domain

import "github.com/pritunl/mongo-go-driver/v2/bson"

const (
	AWS         = "aws"
	Cloudflare  = "cloudflare"
	OracleCloud = "oracle_cloud"

	A     = "A"
	AAAA  = "AAAA"
	CNAME = "CNAME"
	TXT   = "TXT"

	INSERT = "insert"
	UPDATE = "update"
	DELETE = "delete"
)

var (
	Vacant = bson.NilObjectID
)

package authority

import "github.com/pritunl/mongo-go-driver/v2/bson"

const (
	SshKey         = "ssh_key"
	SshCertificate = "ssh_certificate"
)

var (
	Global = bson.NilObjectID
)

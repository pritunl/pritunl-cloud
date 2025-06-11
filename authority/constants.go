package authority

import (
	"github.com/pritunl/mongo-go-driver/bson/primitive"
)

const (
	SshKey         = "ssh_key"
	SshCertificate = "ssh_certificate"
)

var (
	Global = primitive.NilObjectID
)

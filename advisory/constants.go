package advisory

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
)

const (
	RedHat = "rhel"

	Low      = 1
	Medium   = 2
	High     = 3
	Critical = 4

	moderate  = "moderate"
	important = "important"
	critical  = "critical"
)

var (
	Global = bson.NilObjectID
)

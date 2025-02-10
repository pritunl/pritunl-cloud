package spec

import (
	"regexp"

	"github.com/pritunl/mongo-go-driver/bson/primitive"
)

var resourcesRe = regexp.MustCompile("(?s)```yaml(.*?)```")

const (
	All       = "all"
	Icmp      = "icmp"
	Tcp       = "tcp"
	Udp       = "udp"
	Multicast = "multicast"
	Broadcast = "broadcast"

	Private       = "private"
	Private6      = "private6"
	Public        = "public"
	Public6       = "public6"
	OraclePublic  = "oracle_public"
	OraclePrivate = "oracle_private"

	TokenPrefix = "{{"
)

type Base struct {
	Kind string `yaml:"kind"`
}

const (
	Unit = "unit"
)

type Refrence struct {
	Id       primitive.ObjectID `bson:"id" json:"id"`
	Realm    primitive.ObjectID `bson:"realm" json:"realm"`
	Kind     string             `bson:"kind" json:"kind"`
	Selector string             `bson:"selector" json:"selector"`
}

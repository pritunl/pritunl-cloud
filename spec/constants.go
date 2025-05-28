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

	Host          = "host"
	Private       = "private"
	Private6      = "private6"
	Public        = "public"
	Public6       = "public6"
	OraclePublic  = "oracle_public"
	OraclePublic6 = "oracle_public6"
	OraclePrivate = "oracle_private"

	TokenPrefix = "+/"

	Disk     = "disk"
	HostPath = "host_path"
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

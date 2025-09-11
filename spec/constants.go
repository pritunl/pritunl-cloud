package spec

import (
	"regexp"

	"github.com/pritunl/mongo-go-driver/v2/bson"
)

var resourcesRe = regexp.MustCompile("(?s)```yaml(.*?)```")

const (
	All       = "all"
	Icmp      = "icmp"
	Tcp       = "tcp"
	Udp       = "udp"
	Multicast = "multicast"
	Broadcast = "broadcast"

	Host         = "host"
	Private      = "private"
	Private6     = "private6"
	Public       = "public"
	Public6      = "public6"
	CloudPublic  = "cloud_public"
	CloudPublic6 = "cloud_public6"
	CloudPrivate = "cloud_private"

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
	Id       bson.ObjectID `bson:"id" json:"id"`
	Realm    bson.ObjectID `bson:"realm" json:"realm"`
	Kind     string        `bson:"kind" json:"kind"`
	Selector string        `bson:"selector" json:"selector"`
}

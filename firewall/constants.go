package firewall

import "github.com/pritunl/mongo-go-driver/v2/bson"

const (
	All       = "all"
	Icmp      = "icmp"
	Tcp       = "tcp"
	Udp       = "udp"
	Multicast = "multicast"
	Broadcast = "broadcast"
)

var (
	Global = bson.NilObjectID
)

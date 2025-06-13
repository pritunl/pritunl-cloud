package firewall

import (
	"github.com/pritunl/mongo-go-driver/bson/primitive"
)

const (
	All       = "all"
	Icmp      = "icmp"
	Tcp       = "tcp"
	Udp       = "udp"
	Multicast = "multicast"
	Broadcast = "broadcast"
)

var (
	Global = primitive.NilObjectID
)

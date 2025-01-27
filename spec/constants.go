package spec

import (
	"regexp"
)

var resourcesRe = regexp.MustCompile("(?s)```yaml(.*?)```")

const (
	All       = "all"
	Icmp      = "icmp"
	Tcp       = "tcp"
	Udp       = "udp"
	Multicast = "multicast"
	Broadcast = "broadcast"
)

type Base struct {
	Kind string `yaml:"kind"`
}

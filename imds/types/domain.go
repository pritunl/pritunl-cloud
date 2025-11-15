package types

import (
	"net"
)

type Domain struct {
	Domain string `json:"domain"`
	Type   string `json:"type"`
	Ip     net.IP `json:"ip"`
	Target string `json:"target"`
}

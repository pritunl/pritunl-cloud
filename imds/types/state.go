package types

import (
	"net"
	"time"

	"github.com/pritunl/pritunl-cloud/telemetry"
)

type State struct {
	Hash        uint32              `json:"hash"`
	Status      string              `json:"status"`
	Memory      float64             `json:"memory"`
	HugePages   float64             `json:"hugepages"`
	Load1       float64             `json:"load1"`
	Load5       float64             `json:"load5"`
	Load15      float64             `json:"load15"`
	DhcpIp      *net.IPNet          `json:"dhcp_ip"`
	DhcpIp6     *net.IPNet          `json:"dhcp_ip6"`
	DhcpGateway net.IP              `json:"dhcp_gateway"`
	Updates     []*telemetry.Update `json:"updates"`
	Timestamp   time.Time           `json:"timestamp"`
	Output      []*Entry            `json:"output,omitempty"`
}

func (s *State) Final() bool {
	if s.Status == Imaged {
		return true
	}
	return false
}

func (s *State) Copy() *State {
	return &State{
		Hash:        s.Hash,
		Status:      s.Status,
		Memory:      s.Memory,
		HugePages:   s.HugePages,
		Load1:       s.Load1,
		Load5:       s.Load5,
		Load15:      s.Load15,
		DhcpIp:      s.DhcpIp,
		DhcpIp6:     s.DhcpIp6,
		DhcpGateway: s.DhcpGateway,
		Updates:     s.Updates,
		Timestamp:   s.Timestamp,
	}
}

type Entry struct {
	Timestamp time.Time `json:"t"`
	Level     int       `json:"l"`
	Message   string    `json:"m"`
}

const (
	Error = 3
	Info  = 5
)

package types

import (
	"time"
)

type State struct {
	Status    string    `json:"status"`
	Memory    float64   `json:"memory"`
	HugePages float64   `json:"hugepages"`
	Load1     float64   `json:"load1"`
	Load5     float64   `json:"load5"`
	Load15    float64   `json:"load15"`
	Timestamp time.Time `json:"timestamp"`
	Output    []*Entry  `json:"output"`
}

type Entry struct {
	Timestamp time.Time `json:"t"`
	Message   string    `json:"m"`
}

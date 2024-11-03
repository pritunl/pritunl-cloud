package state

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

type Runtimes struct {
	State       time.Duration
	State1      time.Duration
	State2      time.Duration
	State3      time.Duration
	State4      time.Duration
	State5      time.Duration
	State6      time.Duration
	State7      time.Duration
	Network     time.Duration
	Ipset       time.Duration
	Iptables    time.Duration
	Disks       time.Duration
	Instances   time.Duration
	Namespaces  time.Duration
	Services    time.Duration
	Deployments time.Duration
	Imds        time.Duration
	Total       time.Duration
}

func (r *Runtimes) Log() {
	logrus.WithFields(logrus.Fields{
		"state":       fmt.Sprintf("%v", r.State),
		"state1":      fmt.Sprintf("%v", r.State1),
		"state2":      fmt.Sprintf("%v", r.State2),
		"state3":      fmt.Sprintf("%v", r.State3),
		"state4":      fmt.Sprintf("%v", r.State4),
		"state5":      fmt.Sprintf("%v", r.State5),
		"state6":      fmt.Sprintf("%v", r.State6),
		"state7":      fmt.Sprintf("%v", r.State7),
		"network":     fmt.Sprintf("%v", r.Network),
		"ipset":       fmt.Sprintf("%v", r.Ipset),
		"iptables":    fmt.Sprintf("%v", r.Iptables),
		"disks":       fmt.Sprintf("%v", r.Disks),
		"namespaces":  fmt.Sprintf("%v", r.Namespaces),
		"services":    fmt.Sprintf("%v", r.Services),
		"deployments": fmt.Sprintf("%v", r.Deployments),
		"imds":        fmt.Sprintf("%v", r.Imds),
		"total":       fmt.Sprintf("%v", r.Total),
	}).Warn("sync: Excess state sync runtime")
}

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
	State8      time.Duration
	State9      time.Duration
	State10     time.Duration
	State11     time.Duration
	State12     time.Duration
	Network     time.Duration
	Ipset       time.Duration
	Iptables    time.Duration
	Disks       time.Duration
	Instances   time.Duration
	Namespaces  time.Duration
	Pods        time.Duration
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
		"state8":      fmt.Sprintf("%v", r.State8),
		"state9":      fmt.Sprintf("%v", r.State9),
		"state10":     fmt.Sprintf("%v", r.State10),
		"state11":     fmt.Sprintf("%v", r.State11),
		"state12":     fmt.Sprintf("%v", r.State12),
		"network":     fmt.Sprintf("%v", r.Network),
		"ipset":       fmt.Sprintf("%v", r.Ipset),
		"iptables":    fmt.Sprintf("%v", r.Iptables),
		"disks":       fmt.Sprintf("%v", r.Disks),
		"namespaces":  fmt.Sprintf("%v", r.Namespaces),
		"pods":        fmt.Sprintf("%v", r.Pods),
		"deployments": fmt.Sprintf("%v", r.Deployments),
		"imds":        fmt.Sprintf("%v", r.Imds),
		"total":       fmt.Sprintf("%v", r.Total),
	}).Warn("sync: High state sync runtime")
}

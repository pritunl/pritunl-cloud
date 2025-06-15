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
	Wait        time.Duration
	Total       time.Duration
}

func (r *Runtimes) Log() {
	logrus.WithFields(logrus.Fields{
		"network":     fmt.Sprintf("%v", r.Network),
		"ipset":       fmt.Sprintf("%v", r.Ipset),
		"iptables":    fmt.Sprintf("%v", r.Iptables),
		"disks":       fmt.Sprintf("%v", r.Disks),
		"namespaces":  fmt.Sprintf("%v", r.Namespaces),
		"pods":        fmt.Sprintf("%v", r.Pods),
		"deployments": fmt.Sprintf("%v", r.Deployments),
		"imds":        fmt.Sprintf("%v", r.Imds),
		"wait":        fmt.Sprintf("%v", r.Wait),
		"total":       fmt.Sprintf("%v", r.Total),
	}).Warn("sync: High state sync runtime")
}

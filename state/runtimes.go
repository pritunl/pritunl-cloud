package state

import (
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type Runtimes struct {
	State       map[string]time.Duration
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
	lock        sync.Mutex
}

func (r *Runtimes) Init() {
	r.State = map[string]time.Duration{}
}

func (r *Runtimes) SetState(key string, dur time.Duration) {
	r.lock.Lock()
	r.State[key] = dur
	r.lock.Unlock()
}

func (r *Runtimes) Log() {
	fields := logrus.Fields{
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
	}

	for key, dur := range r.State {
		fields[key] = fmt.Sprintf("%v", dur)
	}

	logrus.WithFields(fields).Warn("sync: High state sync runtime")
}

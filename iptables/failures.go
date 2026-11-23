package iptables

import (
	"time"
)

var (
	curFailures = map[string]*Failure{}
)

type Failure struct {
	Count int
	Retry time.Time
}

func (f *Failure) Delay() (delay time.Duration) {
	f.Count += 1

	delay = time.Duration(f.Count-1) * 10 * time.Second
	if delay > 30*time.Second {
		delay = 30 * time.Second
	}
	f.Retry = time.Now().Add(delay)

	return
}

package dhcpc

import (
	"time"
)

type Sleeper struct {
	Interval time.Duration
	Touch    func()
}

func (s *Sleeper) Sleep() {
	remaining := s.Interval

	for remaining > 0 {
		step := min(remaining, TouchInterval)
		time.Sleep(step)
		remaining -= step
		s.Touch()
	}
}

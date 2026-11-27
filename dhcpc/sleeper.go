package dhcpc

import (
	"time"
)

type Sleeper struct {
	Interval time.Duration
	Touch    func()
	retry    time.Duration
}

func (s *Sleeper) sleep(remaining time.Duration) {
	for remaining > 0 {
		step := min(remaining, TouchInterval)
		time.Sleep(step)
		remaining -= step
		s.Touch()
	}
}

func (s *Sleeper) Sleep() {
	s.sleep(s.Interval)
}

func (s *Sleeper) Retry() {
	if s.retry == 0 {
		s.retry = RetryInterval
	}

	s.Touch()
	s.sleep(s.retry)

	s.retry = min(s.retry*2, RetryMaxInterval)
}

func (s *Sleeper) Reset() {
	s.retry = 0
}

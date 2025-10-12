package utils

import (
	"sync"
	"time"
)

type Limiter struct {
	counter int
	limit   int
	lock    sync.Mutex
}

func (l *Limiter) Acquire() (acquired bool) {
	l.lock.Lock()
	if l.counter < l.limit {
		l.counter += 1
		acquired = true
	}
	l.lock.Unlock()
	return
}

func (l *Limiter) Release() {
	l.lock.Lock()
	l.counter -= 1
	if l.counter < 0 {
		panic("limiter: Counter below zero")
	}
	l.lock.Unlock()
}

func NewLimiter(limit int) *Limiter {
	return &Limiter{
		counter: 0,
		limit:   limit,
		lock:    sync.Mutex{},
	}
}

type TimeLimiter struct {
	lastRelease time.Time
	duration    time.Duration
	acquired    bool
	lock        sync.Mutex
}

func (l *TimeLimiter) SetDuration(duration time.Duration) {
	l.lock.Lock()
	l.duration = duration
	l.lock.Unlock()
}

func (l *TimeLimiter) Acquire() (acquired bool) {
	l.lock.Lock()
	defer l.lock.Unlock()

	if l.acquired {
		return false
	}

	if time.Since(l.lastRelease) >= l.duration {
		l.acquired = true
		acquired = true
	}

	return
}

func (l *TimeLimiter) Release() {
	l.lock.Lock()
	defer l.lock.Unlock()

	if !l.acquired {
		panic("limiter: Release called without acquire")
	}

	l.lastRelease = time.Now()
	l.acquired = false
}

func NewTimeLimiter(duration time.Duration) *TimeLimiter {
	return &TimeLimiter{
		lastRelease: time.Time{},
		duration:    duration,
		acquired:    false,
		lock:        sync.Mutex{},
	}
}

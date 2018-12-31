package utils

import "sync"

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
	return
}

func NewLimiter(limit int) *Limiter {
	return &Limiter{
		counter: 0,
		limit:   limit,
		lock:    sync.Mutex{},
	}
}

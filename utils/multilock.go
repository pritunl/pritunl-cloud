package utils

import (
	"sync"
)

type MultiLock struct {
	counts map[string]int
	locks  map[string]*sync.Mutex
	lock   sync.Mutex
}

func (m *MultiLock) Lock(id string) {
	m.lock.Lock()
	val := m.counts[id]
	lock, ok := m.locks[id]
	if !ok {
		lock = &sync.Mutex{}
		m.locks[id] = lock
	}
	m.counts[id] = val + 1
	m.lock.Unlock()

	lock.Lock()
}

func (m *MultiLock) Unlock(id string) {
	m.lock.Lock()
	val := m.counts[id]
	lock := m.locks[id]
	if val <= 1 {
		delete(m.counts, id)
		delete(m.locks, id)
	} else {
		m.counts[id] = val - 1
		lock.Unlock()
	}
	m.lock.Unlock()
}

func (m *MultiLock) Locked(id string) bool {
	m.lock.Lock()
	_, ok := m.locks[id]
	m.lock.Unlock()
	return ok
}

func NewMultiLock() *MultiLock {
	return &MultiLock{
		counts: map[string]int{},
		locks:  map[string]*sync.Mutex{},
		lock:   sync.Mutex{},
	}
}

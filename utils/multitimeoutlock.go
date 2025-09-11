package utils

import (
	"sync"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/constants"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/sirupsen/logrus"
)

type MultiTimeoutLock struct {
	counts    map[string]int
	locks     map[string]*sync.Mutex
	lock      sync.Mutex
	state     map[bson.ObjectID]bool
	stateLock sync.Mutex
	timeout   time.Duration
}

func (m *MultiTimeoutLock) Lock(id string) (lockId bson.ObjectID) {
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

	lockId = bson.NewObjectID()
	m.stateLock.Lock()
	m.state[lockId] = true
	m.stateLock.Unlock()

	if !constants.LockDebug {
		return
	}

	start := time.Now()
	go func() {
		for {
			time.Sleep(1 * time.Second)

			m.stateLock.Lock()
			state := m.state[lockId]
			m.stateLock.Unlock()
			if !state {
				return
			}

			if time.Since(start) > m.timeout {
				err := &errortypes.TimeoutError{
					errors.New("utils: Multi lock timeout"),
				}

				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Error("utils: Multi lock timed out")
				return
			}
		}
	}()

	return
}

func (m *MultiTimeoutLock) LockOpen(id string) (
	acquired bool, lockId bson.ObjectID) {

	m.lock.Lock()
	val := m.counts[id]
	lock, ok := m.locks[id]
	if ok {
		m.lock.Unlock()
		return
	}

	lock = &sync.Mutex{}
	m.locks[id] = lock
	m.counts[id] = val + 1
	m.lock.Unlock()

	acquired = true

	lock.Lock()

	lockId = bson.NewObjectID()
	m.stateLock.Lock()
	m.state[lockId] = true
	m.stateLock.Unlock()

	if !constants.LockDebug {
		return
	}

	start := time.Now()
	go func() {
		for {
			time.Sleep(1 * time.Second)

			m.stateLock.Lock()
			state := m.state[lockId]
			m.stateLock.Unlock()
			if !state {
				return
			}

			if time.Since(start) > m.timeout {
				err := &errortypes.TimeoutError{
					errors.New("utils: Multi lock timeout"),
				}

				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Error("utils: Multi lock timed out")
				return
			}
		}
	}()

	return
}

func (m *MultiTimeoutLock) LockTimeout(id string,
	timeout time.Duration) (lockId bson.ObjectID) {

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

	lockId = bson.NewObjectID()
	m.stateLock.Lock()
	m.state[lockId] = true
	m.stateLock.Unlock()

	if !constants.LockDebug {
		return
	}

	start := time.Now()
	go func() {
		for {
			time.Sleep(1 * time.Second)

			m.stateLock.Lock()
			state := m.state[lockId]
			m.stateLock.Unlock()
			if !state {
				return
			}

			if time.Since(start) > timeout {
				err := &errortypes.TimeoutError{
					errors.New("utils: Multi lock timeout"),
				}

				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Error("utils: Multi lock timed out")
				return
			}
		}
	}()

	return
}

func (m *MultiTimeoutLock) LockOpenTimeout(id string,
	timeout time.Duration) (acquired bool, lockId bson.ObjectID) {

	m.lock.Lock()
	val := m.counts[id]
	lock, ok := m.locks[id]
	if ok {
		m.lock.Unlock()
		return
	}

	lock = &sync.Mutex{}
	m.locks[id] = lock
	m.counts[id] = val + 1
	m.lock.Unlock()

	acquired = true

	lock.Lock()

	lockId = bson.NewObjectID()
	m.stateLock.Lock()
	m.state[lockId] = true
	m.stateLock.Unlock()

	if !constants.LockDebug {
		return
	}

	start := time.Now()
	go func() {
		for {
			time.Sleep(1 * time.Second)

			m.stateLock.Lock()
			state := m.state[lockId]
			m.stateLock.Unlock()
			if !state {
				return
			}

			if time.Since(start) > timeout {
				err := &errortypes.TimeoutError{
					errors.New("utils: Multi lock timeout"),
				}

				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Error("utils: Multi lock timed out")
				return
			}
		}
	}()

	return
}

func (m *MultiTimeoutLock) Unlock(id string, lockId bson.ObjectID) {
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

	m.stateLock.Lock()
	delete(m.state, lockId)
	m.stateLock.Unlock()
}

func (m *MultiTimeoutLock) Locked(id string) bool {
	m.lock.Lock()
	_, ok := m.locks[id]
	m.lock.Unlock()
	return ok
}

func NewMultiTimeoutLock(timeout time.Duration) *MultiTimeoutLock {
	return &MultiTimeoutLock{
		counts:    map[string]int{},
		locks:     map[string]*sync.Mutex{},
		lock:      sync.Mutex{},
		state:     map[bson.ObjectID]bool{},
		stateLock: sync.Mutex{},
		timeout:   timeout,
	}
}

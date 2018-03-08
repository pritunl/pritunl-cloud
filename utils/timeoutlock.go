package utils

import (
	"github.com/Sirupsen/logrus"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"gopkg.in/mgo.v2/bson"
	"sync"
	"time"
)

type TimeoutLock struct {
	lock      sync.Mutex
	state     map[bson.ObjectId]bool
	stateLock sync.Mutex
	timeout   time.Duration
}

func (l *TimeoutLock) Lock() (id bson.ObjectId) {
	id = bson.NewObjectId()
	l.lock.Lock()

	start := time.Now()
	err := &errortypes.TimeoutError{
		errors.New("utils: Lock timeout"),
	}

	l.stateLock.Lock()
	l.state[id] = true
	l.stateLock.Unlock()

	go func() {
		for {
			time.Sleep(1 * time.Second)

			l.stateLock.Lock()
			state := l.state[id]
			l.stateLock.Unlock()
			if !state {
				return
			}

			if time.Since(start) > l.timeout {
				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Error("utils: Lock timed out")
				return
			}
		}
	}()

	return
}

func (l *TimeoutLock) Unlock(id bson.ObjectId) {
	l.lock.Unlock()
	l.stateLock.Lock()
	delete(l.state, id)
	l.stateLock.Unlock()
}

func NewTimeoutLock(timeout time.Duration) *TimeoutLock {
	return &TimeoutLock{
		lock:      sync.Mutex{},
		state:     map[bson.ObjectId]bool{},
		stateLock: sync.Mutex{},
		timeout:   timeout,
	}
}

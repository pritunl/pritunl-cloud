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

type TimeoutLock struct {
	lock      sync.Mutex
	state     map[bson.ObjectID]bool
	stateLock sync.Mutex
	timeout   time.Duration
}

func (l *TimeoutLock) Lock() (id bson.ObjectID) {
	id = bson.NewObjectID()
	l.lock.Lock()

	l.stateLock.Lock()
	l.state[id] = true
	l.stateLock.Unlock()

	if !constants.LockDebug {
		return
	}

	start := time.Now()
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
				err := &errortypes.TimeoutError{
					errors.New("utils: Multi lock timeout"),
				}

				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Error("utils: Lock timed out")
				return
			}
		}
	}()

	return
}

func (l *TimeoutLock) Unlock(id bson.ObjectID) {
	l.lock.Unlock()
	l.stateLock.Lock()
	delete(l.state, id)
	l.stateLock.Unlock()
}

func NewTimeoutLock(timeout time.Duration) *TimeoutLock {
	return &TimeoutLock{
		lock:      sync.Mutex{},
		state:     map[bson.ObjectID]bool{},
		stateLock: sync.Mutex{},
		timeout:   timeout,
	}
}

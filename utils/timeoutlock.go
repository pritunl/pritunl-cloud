package utils

import (
	"github.com/Sirupsen/logrus"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/constants"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"sync"
	"time"
)

type TimeoutLock struct {
	lock      sync.Mutex
	state     map[primitive.ObjectID]bool
	stateLock sync.Mutex
	timeout   time.Duration
}

func (l *TimeoutLock) Lock() (id primitive.ObjectID) {
	id = primitive.NewObjectID()
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

func (l *TimeoutLock) Unlock(id primitive.ObjectID) {
	l.lock.Unlock()
	l.stateLock.Lock()
	delete(l.state, id)
	l.stateLock.Unlock()
}

func NewTimeoutLock(timeout time.Duration) *TimeoutLock {
	return &TimeoutLock{
		lock:      sync.Mutex{},
		state:     map[primitive.ObjectID]bool{},
		stateLock: sync.Mutex{},
		timeout:   timeout,
	}
}

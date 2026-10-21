package lock

import (
	"fmt"
	"sync"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/sirupsen/logrus"
)

const (
	lvmRefreshInterval = 30 * time.Second
)

type LvmLockedError struct {
	errors.DropboxError
}

type LvmLocker struct {
	Id        string        `bson:"_id" json:"_id"`
	Node      bson.ObjectID `bson:"node" json:"node"`
	Timestamp time.Time     `bson:"timestamp" json:"timestamp"`
	lock      sync.Mutex
	stop      chan struct{}
}

func NewLvmLocker(vgName, lvName string) (lckr *LvmLocker) {
	lckr = &LvmLocker{
		Id:   fmt.Sprintf("%s/%s", vgName, lvName),
		Node: node.Self.Id,
	}

	return
}

func (l *LvmLocker) Lock(db *database.Database) (acquired bool, err error) {
	coll := db.LvmLock()

	l.Timestamp = time.Now()

	_, err = coll.InsertOne(db, l)
	if err != nil {
		err = database.ParseError(err)
		if _, ok := err.(*database.DuplicateKeyError); ok {
			err = nil
			return
		}
		return
	}

	acquired = true
	l.Start()

	return
}

func (l *LvmLocker) Refresh(db *database.Database) (err error) {
	coll := db.LvmLock()

	_, err = coll.UpdateOne(db, &bson.M{
		"_id":  l.Id,
		"node": l.Node,
	}, &bson.M{
		"$set": &bson.M{
			"timestamp": time.Now(),
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func (l *LvmLocker) Unlock(db *database.Database) (err error) {
	coll := db.LvmLock()

	l.Stop()

	_, err = coll.DeleteOne(db, &bson.M{
		"_id":  l.Id,
		"node": l.Node,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func (l *LvmLocker) run(stop chan struct{}) {
	for {
		select {
		case <-stop:
			return
		case <-time.After(lvmRefreshInterval):
		}

		db := database.GetDatabase()
		err := l.Refresh(db)
		db.Close()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"lock":  l.Id,
				"error": err,
			}).Warning("lock: Failed to refresh lvm lock")
		}
	}
}

func (l *LvmLocker) Start() {
	l.lock.Lock()
	defer l.lock.Unlock()

	if l.stop != nil {
		return
	}

	l.stop = make(chan struct{})
	go l.run(l.stop)
}

func (l *LvmLocker) Stop() {
	l.lock.Lock()
	defer l.lock.Unlock()

	if l.stop == nil {
		return
	}

	close(l.stop)
	l.stop = nil
}

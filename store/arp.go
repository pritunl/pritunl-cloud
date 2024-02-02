package store

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"sync"
	"time"
)

var (
	arpStores     = map[primitive.ObjectID]ArpStore{}
	arpStoresLock = sync.Mutex{}
)

type ArpStore struct {
	Records   set.Set
	Timestamp time.Time
}

func GetArp(instId primitive.ObjectID) (arpStore ArpStore, ok bool) {
	arpStoresLock.Lock()
	arpStore, ok = arpStores[instId]
	arpStoresLock.Unlock()

	if ok {
		arpStore.Records = arpStore.Records.Copy()
	}

	return
}

func SetArp(instId primitive.ObjectID, records set.Set) {
	arpStoresLock.Lock()
	arpStores[instId] = ArpStore{
		Records:   records.Copy(),
		Timestamp: time.Now(),
	}
	arpStoresLock.Unlock()
}

func RemArp(instId primitive.ObjectID) {
	arpStoresLock.Lock()
	delete(arpStores, instId)
	arpStoresLock.Unlock()
}

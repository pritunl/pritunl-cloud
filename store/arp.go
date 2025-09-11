package store

import (
	"sync"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/v2/bson"
)

var (
	arpStores     = map[bson.ObjectID]ArpStore{}
	arpStoresLock = sync.Mutex{}
)

type ArpStore struct {
	Records   set.Set
	Timestamp time.Time
}

func GetArp(instId bson.ObjectID) (arpStore ArpStore, ok bool) {
	arpStoresLock.Lock()
	arpStore, ok = arpStores[instId]
	arpStoresLock.Unlock()

	if ok {
		arpStore.Records = arpStore.Records.Copy()
	}

	return
}

func SetArp(instId bson.ObjectID, records set.Set) {
	arpStoresLock.Lock()
	arpStores[instId] = ArpStore{
		Records:   records.Copy(),
		Timestamp: time.Now(),
	}
	arpStoresLock.Unlock()
}

func RemArp(instId bson.ObjectID) {
	arpStoresLock.Lock()
	delete(arpStores, instId)
	arpStoresLock.Unlock()
}

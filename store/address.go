package store

import (
	"gopkg.in/mgo.v2/bson"
	"sync"
	"time"
)

var (
	addressStores     = map[bson.ObjectId]AddressStore{}
	addressStoresLock = sync.Mutex{}
)

type AddressStore struct {
	Addr      string
	Addr6     string
	Timestamp time.Time
}

func GetAddress(virtId bson.ObjectId) (addressStore AddressStore, ok bool) {
	addressStoresLock.Lock()
	addressStore, ok = addressStores[virtId]
	addressStoresLock.Unlock()

	return
}

func SetAddress(virtId bson.ObjectId, addr, addr6 string) {
	addressStoresLock.Lock()
	addressStores[virtId] = AddressStore{
		Addr:      addr,
		Addr6:     addr6,
		Timestamp: time.Now(),
	}
	addressStoresLock.Unlock()
}

func RemAddress(addressId bson.ObjectId) {
	addressStoresLock.Lock()
	delete(addressStores, addressId)
	addressStoresLock.Unlock()
}

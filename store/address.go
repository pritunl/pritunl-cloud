package store

import (
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"sync"
	"time"
)

var (
	addressStores     = map[primitive.ObjectID]AddressStore{}
	addressStoresLock = sync.Mutex{}
)

type AddressStore struct {
	Addr      string
	Addr6     string
	Timestamp time.Time
}

func GetAddress(virtId primitive.ObjectID) (addressStore AddressStore, ok bool) {
	addressStoresLock.Lock()
	addressStore, ok = addressStores[virtId]
	addressStoresLock.Unlock()

	return
}

func SetAddress(virtId primitive.ObjectID, addr, addr6 string) {
	addressStoresLock.Lock()
	addressStores[virtId] = AddressStore{
		Addr:      addr,
		Addr6:     addr6,
		Timestamp: time.Now(),
	}
	addressStoresLock.Unlock()
}

func RemAddress(addressId primitive.ObjectID) {
	addressStoresLock.Lock()
	delete(addressStores, addressId)
	addressStoresLock.Unlock()
}

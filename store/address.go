package store

import (
	"sync"
	"time"

	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/settings"
)

var (
	addressStores     = map[primitive.ObjectID]*AddressStore{}
	addressStoresLock = sync.Mutex{}
)

type AddressStore struct {
	Addr      string
	Addr6     string
	Timestamp time.Time
	Refresh   time.Duration
}

func GetAddress(virtId primitive.ObjectID) (
	addressStore *AddressStore, ok bool) {

	addressStoresLock.Lock()
	addressStore, ok = addressStores[virtId]
	addressStoresLock.Unlock()

	ttl := settings.Hypervisor.AddressRefreshTtl
	if ok && ttl != 0 && time.Since(addressStore.Timestamp) > time.Duration(
		ttl)*time.Second && node.Self.IsDhcp6() {

		ok = false
	}

	return
}

func SetAddress(virtId primitive.ObjectID, addr, addr6 string) {
	addressStoresLock.Lock()
	now := time.Now()

	addressStore := addressStores[virtId]
	if addressStore != nil && addressStore.Refresh != 0 {
		refreshTtl := time.Duration(
			settings.Hypervisor.AddressRefreshTtl) * time.Second
		now = now.Add(-refreshTtl).Add(addressStore.Refresh)
	}

	addressStores[virtId] = &AddressStore{
		Addr:      addr,
		Addr6:     addr6,
		Timestamp: now,
	}

	addressStoresLock.Unlock()
}

func SetAddressExpire(virtId primitive.ObjectID, ttl time.Duration) {
	addressStoresLock.Lock()
	addressStore, ok := addressStores[virtId]
	if ok {
		refreshTtl := time.Duration(
			settings.Hypervisor.AddressRefreshTtl) * time.Second
		addressStore.Timestamp = time.Now().Add(-refreshTtl).Add(ttl)
	}
	addressStoresLock.Unlock()
}

func SetAddressExpireMulti(virtId primitive.ObjectID,
	ttl, ttl2 time.Duration) {

	addressStoresLock.Lock()
	addressStore, ok := addressStores[virtId]
	if ok {
		refreshTtl := time.Duration(
			settings.Hypervisor.AddressRefreshTtl) * time.Second
		addressStore.Timestamp = time.Now().Add(-refreshTtl).Add(ttl)
		addressStore.Refresh = ttl2
	}
	addressStoresLock.Unlock()
}

func RemAddress(addressId primitive.ObjectID) {
	addressStoresLock.Lock()
	delete(addressStores, addressId)
	addressStoresLock.Unlock()
}

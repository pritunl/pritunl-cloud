package store

import (
	"sync"
	"time"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/vm"
)

var (
	disksStores     = map[bson.ObjectID]DisksStore{}
	disksStoresLock = sync.Mutex{}
)

type DisksStore struct {
	Disks     []vm.Disk
	Timestamp time.Time
}

func GetDisks(virtId bson.ObjectID) (disksStore DisksStore, ok bool) {
	disksStoresLock.Lock()
	disksStore, ok = disksStores[virtId]
	disksStoresLock.Unlock()

	if ok {
		disksStore.Disks = append([]vm.Disk{}, disksStore.Disks...)
	}

	return
}

func SetDisks(virtId bson.ObjectID, disks []*vm.Disk) {
	disksRef := []vm.Disk{}
	for _, dsk := range disks {
		disksRef = append(disksRef, *dsk)
	}

	disksStoresLock.Lock()
	disksStores[virtId] = DisksStore{
		Disks:     disksRef,
		Timestamp: time.Now(),
	}
	disksStoresLock.Unlock()
}

func RemDisks(virtId bson.ObjectID) {
	disksStoresLock.Lock()
	delete(disksStores, virtId)
	disksStoresLock.Unlock()
}

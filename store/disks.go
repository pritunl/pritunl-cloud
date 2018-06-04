package store

import (
	"github.com/pritunl/pritunl-cloud/vm"
	"gopkg.in/mgo.v2/bson"
	"sync"
	"time"
)

var (
	disksStores     = map[bson.ObjectId]DisksStore{}
	disksStoresLock = sync.Mutex{}
)

type DisksStore struct {
	Disks     []vm.Disk
	Timestamp time.Time
}

func GetDisks(virtId bson.ObjectId) (disksStore DisksStore, ok bool) {
	disksStoresLock.Lock()
	disksStore, ok = disksStores[virtId]
	disksStoresLock.Unlock()

	if ok {
		disksStore.Disks = append([]vm.Disk{}, disksStore.Disks...)
	}

	return
}

func SetDisks(virtId bson.ObjectId, disks []*vm.Disk) {
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

func RemDisks(virtId bson.ObjectId) {
	disksStoresLock.Lock()
	delete(disksStores, virtId)
	disksStoresLock.Unlock()
}

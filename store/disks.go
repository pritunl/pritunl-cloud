package store

import (
	"sync"
	"time"

	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/vm"
)

var (
	disksStores     = map[primitive.ObjectID]DisksStore{}
	disksStoresLock = sync.Mutex{}
)

type DisksStore struct {
	Disks     []vm.Disk
	Timestamp time.Time
}

func GetDisks(virtId primitive.ObjectID) (disksStore DisksStore, ok bool) {
	disksStoresLock.Lock()
	disksStore, ok = disksStores[virtId]
	disksStoresLock.Unlock()

	if ok {
		disksStore.Disks = append([]vm.Disk{}, disksStore.Disks...)
	}

	return
}

func SetDisks(virtId primitive.ObjectID, disks []*vm.Disk) {
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

func RemDisks(virtId primitive.ObjectID) {
	disksStoresLock.Lock()
	delete(disksStores, virtId)
	disksStoresLock.Unlock()
}

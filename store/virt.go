package store

import (
	"sync"
	"time"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/vm"
)

var (
	virtStores     = map[bson.ObjectID]VirtStore{}
	virtStoresLock = sync.Mutex{}
)

type VirtStore struct {
	Virt      vm.VirtualMachine
	Timestamp time.Time
}

func GetVirt(virtId bson.ObjectID) (virtStore VirtStore, ok bool) {
	virtStoresLock.Lock()
	virtStore, ok = virtStores[virtId]
	virtStoresLock.Unlock()

	return
}

func SetVirt(virtId bson.ObjectID, virt *vm.VirtualMachine) {
	virtRef := *virt
	virtRef.Disks = nil

	virtStoresLock.Lock()
	virtStores[virtId] = VirtStore{
		Virt:      virtRef,
		Timestamp: time.Now(),
	}
	virtStoresLock.Unlock()
}

func RemVirt(virtId bson.ObjectID) {
	virtStoresLock.Lock()
	delete(virtStores, virtId)
	virtStoresLock.Unlock()
}

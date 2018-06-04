package store

import (
	"github.com/pritunl/pritunl-cloud/vm"
	"gopkg.in/mgo.v2/bson"
	"sync"
	"time"
)

var (
	virtStores     = map[bson.ObjectId]VirtStore{}
	virtStoresLock = sync.Mutex{}
)

type VirtStore struct {
	Virt      vm.VirtualMachine
	Timestamp time.Time
}

func GetVirt(virtId bson.ObjectId) (virtStore VirtStore, ok bool) {
	virtStoresLock.Lock()
	virtStore, ok = virtStores[virtId]
	virtStoresLock.Unlock()

	return
}

func SetVirt(virtId bson.ObjectId, virt *vm.VirtualMachine) {
	virtRef := *virt
	virtRef.Disks = nil

	virtStoresLock.Lock()
	virtStores[virtId] = VirtStore{
		Virt:      virtRef,
		Timestamp: time.Now(),
	}
	virtStoresLock.Unlock()
}

func RemVirt(virtId bson.ObjectId) {
	virtStoresLock.Lock()
	delete(virtStores, virtId)
	virtStoresLock.Unlock()
}

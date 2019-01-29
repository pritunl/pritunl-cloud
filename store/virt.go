package store

import (
	"sync"
	"time"

	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/vm"
)

var (
	virtStores     = map[primitive.ObjectID]VirtStore{}
	virtStoresLock = sync.Mutex{}
)

type VirtStore struct {
	Virt      vm.VirtualMachine
	Timestamp time.Time
}

func GetVirt(virtId primitive.ObjectID) (virtStore VirtStore, ok bool) {
	virtStoresLock.Lock()
	virtStore, ok = virtStores[virtId]
	virtStoresLock.Unlock()

	return
}

func SetVirt(virtId primitive.ObjectID, virt *vm.VirtualMachine) {
	virtRef := *virt
	virtRef.Disks = nil

	virtStoresLock.Lock()
	virtStores[virtId] = VirtStore{
		Virt:      virtRef,
		Timestamp: time.Now(),
	}
	virtStoresLock.Unlock()
}

func RemVirt(virtId primitive.ObjectID) {
	virtStoresLock.Lock()
	delete(virtStores, virtId)
	virtStoresLock.Unlock()
}

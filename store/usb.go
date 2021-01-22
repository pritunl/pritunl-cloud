package store

import (
	"sync"
	"time"

	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/vm"
)

var (
	usbsStores     = map[primitive.ObjectID]UsbsStore{}
	usbsStoresLock = sync.Mutex{}
)

type UsbsStore struct {
	Usbs      []vm.UsbDevice
	Timestamp time.Time
}

func GetUsbs(virtId primitive.ObjectID) (usbsStore UsbsStore, ok bool) {
	usbsStoresLock.Lock()
	usbsStore, ok = usbsStores[virtId]
	usbsStoresLock.Unlock()

	if ok {
		usbsStore.Usbs = append([]vm.UsbDevice{}, usbsStore.Usbs...)
	}

	return
}

func SetUsbs(virtId primitive.ObjectID, usbs []*vm.UsbDevice) {
	usbsRef := []vm.UsbDevice{}
	for _, dsk := range usbs {
		usbsRef = append(usbsRef, *dsk)
	}

	usbsStoresLock.Lock()
	usbsStores[virtId] = UsbsStore{
		Usbs:      usbsRef,
		Timestamp: time.Now(),
	}
	usbsStoresLock.Unlock()
}

func RemUsbs(virtId primitive.ObjectID) {
	usbsStoresLock.Lock()
	delete(usbsStores, virtId)
	usbsStoresLock.Unlock()
}

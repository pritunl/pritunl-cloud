package store

import (
	"github.com/pritunl/pritunl-cloud/vpc"
	"gopkg.in/mgo.v2/bson"
	"sync"
	"time"
)

var (
	routesStores     = map[bson.ObjectId]RoutesStore{}
	routesStoresLock = sync.Mutex{}
)

type RoutesStore struct {
	Routes    []vpc.Route
	Routes6   []vpc.Route
	Timestamp time.Time
}

func GetRoutes(instId bson.ObjectId) (routesStore RoutesStore, ok bool) {
	routesStoresLock.Lock()
	routesStore, ok = routesStores[instId]
	routesStoresLock.Unlock()

	if ok {
		routesStore.Routes = append([]vpc.Route{}, routesStore.Routes...)
	}

	return
}

func SetRoutes(instId bson.ObjectId, routes, routes6 []vpc.Route) {
	routesStoresLock.Lock()
	routesStores[instId] = RoutesStore{
		Routes:    append([]vpc.Route{}, routes...),
		Routes6:   append([]vpc.Route{}, routes6...),
		Timestamp: time.Now(),
	}
	routesStoresLock.Unlock()
}

func RemRoutes(instId bson.ObjectId) {
	routesStoresLock.Lock()
	delete(routesStores, instId)
	routesStoresLock.Unlock()
}

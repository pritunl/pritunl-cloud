package store

import (
	"sync"
	"time"

	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/vpc"
)

var (
	routesStores     = map[primitive.ObjectID]RoutesStore{}
	routesStoresLock = sync.Mutex{}
)

type RoutesStore struct {
	IcmpRedirects bool
	Routes        []vpc.Route
	Routes6       []vpc.Route
	Timestamp     time.Time
}

func GetRoutes(instId primitive.ObjectID) (routesStore RoutesStore, ok bool) {
	routesStoresLock.Lock()
	routesStore, ok = routesStores[instId]
	routesStoresLock.Unlock()

	if ok {
		routesStore.Routes = append([]vpc.Route{}, routesStore.Routes...)
	}

	return
}

func SetRoutes(instId primitive.ObjectID, icmpRedirects bool,
	routes, routes6 []vpc.Route) {

	routesStoresLock.Lock()
	routesStores[instId] = RoutesStore{
		IcmpRedirects: icmpRedirects,
		Routes:        append([]vpc.Route{}, routes...),
		Routes6:       append([]vpc.Route{}, routes6...),
		Timestamp:     time.Now(),
	}
	routesStoresLock.Unlock()
}

func RemRoutes(instId primitive.ObjectID) {
	routesStoresLock.Lock()
	delete(routesStores, instId)
	routesStoresLock.Unlock()
}

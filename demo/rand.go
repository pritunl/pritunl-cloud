package demo

import (
	"sync"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/utils"
)

var (
	lock            = sync.Mutex{}
	ipStore         = map[bson.ObjectID]string{}
	ip6Store        = map[bson.ObjectID]string{}
	privateIpStore  = map[bson.ObjectID]string{}
	privateIp6Store = map[bson.ObjectID]string{}
)

func RandIp(instId bson.ObjectID) (addr string) {
	lock.Lock()
	defer lock.Unlock()

	addr = ipStore[instId]
	if addr == "" {
		addr = utils.RandIp()
		ipStore[instId] = addr
	}

	return
}

func RandIp6(instId bson.ObjectID) (addr string) {
	lock.Lock()
	defer lock.Unlock()

	addr = ip6Store[instId]
	if addr == "" {
		addr = utils.RandIp6()
		ip6Store[instId] = addr
	}

	return
}

func RandPrivateIp(instId bson.ObjectID) (addr string) {
	lock.Lock()
	defer lock.Unlock()

	addr = privateIpStore[instId]
	if addr == "" {
		addr = utils.RandPrivateIp()
		privateIpStore[instId] = addr
	}

	return
}

func RandPrivateIp6(instId bson.ObjectID) (addr string) {
	lock.Lock()
	defer lock.Unlock()

	addr = privateIp6Store[instId]
	if addr == "" {
		addr = utils.RandPrivateIp6()
		privateIp6Store[instId] = addr
	}

	return
}

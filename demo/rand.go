package demo

import (
	"github.com/pritunl/pritunl-cloud/utils"
	"gopkg.in/mgo.v2/bson"
	"sync"
)

var (
	lock     = sync.Mutex{}
	ipStore  = map[bson.ObjectId]string{}
	ip6Store = map[bson.ObjectId]string{}
)

func RandIp(instId bson.ObjectId) (addr string) {
	lock.Lock()
	defer lock.Unlock()

	addr = ipStore[instId]
	if addr == "" {
		addr = utils.RandIp()
		ipStore[instId] = addr
	}

	return
}

func RandIp6(instId bson.ObjectId) (addr string) {
	lock.Lock()
	defer lock.Unlock()

	addr = ip6Store[instId]
	if addr == "" {
		addr = utils.RandIp6()
		ip6Store[instId] = addr
	}

	return
}

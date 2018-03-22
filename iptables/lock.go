package iptables

import (
	"sync"
)

var lock = sync.Mutex{}

func Lock() {
	lock.Lock()
}

func Unlock() {
	lock.Unlock()
}

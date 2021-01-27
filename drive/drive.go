package drive

import (
	"sync"
	"time"
)

var (
	syncLast  time.Time
	syncLock  sync.Mutex
	syncCache []*Device
)

type Device struct {
	Id string `bson:"id" json:"id"`
}

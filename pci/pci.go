package pci

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
	Slot   string `bson:"slot" json:"slot"`
	Class  string `bson:"class" json:"class"`
	Name   string `bson:"name" json:"name"`
	Driver string `bson:"driver" json:"driver"`
}

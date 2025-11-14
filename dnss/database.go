package dnss

import (
	"net"
	"sync/atomic"
)

var (
	database atomic.Pointer[Database]
)

type Database struct {
	A     map[string][]net.IP `json:"a"`
	AAAA  map[string][]net.IP `json:"aaaa"`
	CNAME map[string]string   `json:"cname"`
}

func init() {
	database.Store(&Database{
		A:     map[string][]net.IP{},
		AAAA:  map[string][]net.IP{},
		CNAME: map[string]string{},
	})
}

func UpdateDatabase(db *Database) {
	database.Store(db)
}

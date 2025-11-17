package dnss

import (
	"net"
	"sync/atomic"

	"github.com/pritunl/pritunl-cloud/domain"
	"github.com/pritunl/pritunl-cloud/imds/types"
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

func LoadConfig(domains []*types.Domain) {
	db := &Database{
		A:     map[string][]net.IP{},
		AAAA:  map[string][]net.IP{},
		CNAME: map[string]string{},
	}

	for _, domn := range domains {
		switch domn.Type {
		case domain.A:
			db.A[domn.Domain] = append(db.A[domn.Domain], domn.Ip)
		case domain.AAAA:
			db.AAAA[domn.Domain] = append(db.AAAA[domn.Domain], domn.Ip)
		case domain.CNAME:
			db.CNAME[domn.Domain] = domn.Target
		}
	}

	UpdateDatabase(db)
}

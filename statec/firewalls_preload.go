package statec

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
)

var (
	FirewallsPreload    = &FirewallsPreloadState{}
	FirewallsPreloadPkg = NewPackage(FirewallsPreload)
	firewallsCache      = NewCache(
		"firewalls",
		nil,
		func() Scope {
			return Scope{
				Filter: bson.M{},
				Key:    "all",
				Ok:     true,
			}
		},
		nil,
	)
)

type FirewallsPreloadState struct {
}

func (p *FirewallsPreloadState) Refresh(pkg *Package,
	db *database.Database) (err error) {

	err = firewallsCache.Sync(db)
	if err != nil {
		return
	}

	return
}

func (p *FirewallsPreloadState) Apply(st *State) {
}

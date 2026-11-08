package statec

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
)

var (
	AuthoritiesPreload    = &AuthoritiesPreloadState{}
	AuthoritiesPreloadPkg = NewPackage(AuthoritiesPreload)
	authoritiesCache      = NewCache(
		"authorities",
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

type AuthoritiesPreloadState struct {
}

func (p *AuthoritiesPreloadState) Refresh(pkg *Package,
	db *database.Database) (err error) {

	err = authoritiesCache.Sync(db)
	if err != nil {
		return
	}

	return
}

func (p *AuthoritiesPreloadState) Apply(st *State) {
}

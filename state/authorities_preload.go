package state

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/pritunl-cloud/authority"
	"github.com/pritunl/pritunl-cloud/database"
)

var (
	AuthoritiesPreload    = &AuthoritiesPreloadState{}
	AuthoritiesPreloadPkg = NewPackage(AuthoritiesPreload)
)

type AuthoritiesPreloadState struct {
	authoritiesMap      map[string][]*authority.Authority
	authoritiesRolesSet set.Set
}

func (p *AuthoritiesPreloadState) Authorities() map[string][]*authority.Authority {
	return p.authoritiesMap
}

func (p *AuthoritiesPreloadState) RolesSet() set.Set {
	return p.authoritiesRolesSet
}

func (p *AuthoritiesPreloadState) Refresh(pkg *Package,
	db *database.Database) (err error) {

	roles, rolesSet := InstancesPreload.GetRoles()
	if len(roles) == 0 {
		p.authoritiesMap = map[string][]*authority.Authority{}
		p.authoritiesRolesSet = set.NewSet()
		return
	}

	authrsMap, err := authority.GetMapRoles(db, &bson.M{
		"roles": &bson.M{
			"$in": roles,
		},
	})
	if err != nil {
		return
	}
	p.authoritiesMap = authrsMap
	p.authoritiesRolesSet = rolesSet

	return
}

func (p *AuthoritiesPreloadState) Apply(st *State) {
}

package state

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/authority"
	"github.com/pritunl/pritunl-cloud/database"
)

var (
	Authorities    = &AuthoritiesState{}
	AuthoritiesPkg = NewPackage(Authorities)
)

type AuthoritiesState struct {
	authoritiesMap map[string][]*authority.Authority
}

func (p *AuthoritiesState) GetInstaceAuthorities(
	orgId primitive.ObjectID, roles []string) []*authority.Authority {

	authrSet := set.NewSet()
	authrs := []*authority.Authority{}

	for _, role := range roles {
		for _, authr := range p.authoritiesMap[role] {
			if authrSet.Contains(authr.Id) || authr.Organization != orgId {
				continue
			}
			authrSet.Add(authr.Id)
			authrs = append(authrs, authr)
		}
	}

	return authrs
}

func (p *AuthoritiesState) Refresh(pkg *Package,
	db *database.Database) (err error) {

	_, rolesSet := InstancesPreload.GetRoles()
	authorities := AuthoritiesPreload.Authorities()
	preloadRolesSet := AuthoritiesPreload.RolesSet()
	roles := rolesSet.Copy()
	roles.Subtract(preloadRolesSet)

	missRoles := []string{}
	for roleInf := range roles.Iter() {
		missRoles = append(missRoles, roleInf.(string))
	}

	if len(missRoles) > 0 {
		missAuthorities, e := authority.GetMapRoles(db, &bson.M{
			"roles": &bson.M{
				"$in": missRoles,
			},
		})
		if e != nil {
			err = e
			return
		}

		for role, authrs := range missAuthorities {
			authorities[role] = authrs
		}
	}

	p.authoritiesMap = authorities

	return
}

func (p *AuthoritiesState) Apply(st *State) {
	st.GetInstaceAuthorities = p.GetInstaceAuthorities
}

func init() {
	AuthoritiesPkg.
		After(AuthoritiesPreload).
		After(InstancesPreload)
}

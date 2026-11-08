package statec

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/v2/bson"
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
	orgId bson.ObjectID, roles []string) []*authority.Authority {

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
	authorities := map[string][]*authority.Authority{}

	if rolesSet != nil && rolesSet.Len() > 0 {
		authrs, e := Decode[authority.Authority](authoritiesCache)
		if e != nil {
			err = e
			return
		}

		for _, authr := range authrs {
			matched := false
			for _, role := range authr.Roles {
				if rolesSet.Contains(role) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}

			for _, role := range authr.Roles {
				roleAuthrs := authorities[role]
				if roleAuthrs == nil {
					roleAuthrs = []*authority.Authority{}
				}
				authorities[role] = append(roleAuthrs, authr)
			}
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

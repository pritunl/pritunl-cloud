package statec

import (
	"net"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/domain"
	"github.com/pritunl/pritunl-cloud/imds/types"
)

var (
	Domains      = &DomainsState{}
	DomainsPkg   = NewPackage(Domains)
	domainsCache = NewCache(
		"domains",
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
	domainsRecordsCache = NewCache(
		"domains_records",
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

type DomainsState struct {
	domains map[bson.ObjectID][]*types.Domain
}

func (p *DomainsState) GetDomains(orgId bson.ObjectID) []*types.Domain {
	return p.domains[orgId]
}

func (p *DomainsState) Refresh(pkg *Package,
	db *database.Database) (err error) {

	rootDomains := map[bson.ObjectID]*domain.Domain{}
	records := map[bson.ObjectID][]*types.Domain{}

	err = domainsCache.Sync(db)
	if err != nil {
		return
	}

	domains, err := Decode[domain.Domain](domainsCache)
	if err != nil {
		return
	}

	for _, dmn := range domains {
		rootDomains[dmn.Id] = dmn
	}

	err = domainsRecordsCache.Sync(db)
	if err != nil {
		return
	}

	recs, err := Decode[domain.Record](domainsRecordsCache)
	if err != nil {
		return
	}

	for _, rec := range recs {
		if rec.IsDeleted() {
			continue
		}

		dmn := rootDomains[rec.Domain]
		if dmn == nil {
			continue
		}

		dmnRec := &types.Domain{
			Domain: rec.SubDomain + "." + dmn.RootDomain + ".",
			Type:   rec.Type,
		}

		switch rec.Type {
		case domain.A:
			dmnRec.Ip = net.ParseIP(rec.Value)
		case domain.AAAA:
			dmnRec.Ip = net.ParseIP(rec.Value)
		case domain.CNAME:
			dmnRec.Target = rec.Value + "."
		default:
			continue
		}

		records[dmn.Organization] = append(records[dmn.Organization], dmnRec)
	}

	p.domains = records

	return
}

func (p *DomainsState) Apply(st *State) {
	st.GetDomains = p.GetDomains
}

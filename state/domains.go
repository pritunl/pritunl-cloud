package state

import (
	"net"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/mongo-go-driver/v2/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/domain"
	"github.com/pritunl/pritunl-cloud/imds/types"
)

var (
	Domains    = &DomainsState{}
	DomainsPkg = NewPackage(Domains)
)

type DomainsState struct {
	domains map[bson.ObjectID][]*types.Domain
}

func (p *DomainsState) GetDomains(orgId bson.ObjectID) []*types.Domain {
	return p.domains[orgId]
}

func (p *DomainsState) Refresh(pkg *Package,
	db *database.Database) (err error) {

	coll := db.Domains()
	rootDomains := map[bson.ObjectID]*domain.Domain{}
	records := map[bson.ObjectID][]*types.Domain{}

	cursor, err := coll.Find(db, bson.M{})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		dmn := &domain.Domain{}
		err = cursor.Decode(dmn)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		rootDomains[dmn.Id] = dmn
	}

	coll = db.DomainsRecords()

	cursor, err = coll.Find(
		db,
		bson.M{},
		options.Find().SetSort(bson.D{{"_id", 1}}),
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		rec := &domain.Record{}
		err = cursor.Decode(rec)
		if err != nil {
			err = database.ParseError(err)
			return
		}

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

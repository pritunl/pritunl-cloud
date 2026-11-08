package statec

import (
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/datacenter"
	"github.com/pritunl/pritunl-cloud/node"
)

var (
	Datacenter      = &DatacenterState{}
	DatacenterPkg   = NewPackage(Datacenter)
	datacenterCache = NewCache(
		"datacenters",
		nil,
		func() Scope {
			dcId := node.Self.Datacenter
			if dcId.IsZero() {
				return Scope{}
			}

			return Scope{
				Filter: bson.M{
					"_id": dcId,
				},
				Key: dcId.Hex(),
				Ok:  true,
			}
		},
		func(evt *Event) bool {
			dcId, ok := evt.DocId()
			return ok && dcId == node.Self.Datacenter
		},
	)
)

type DatacenterState struct {
	nodeDatacenter *datacenter.Datacenter
}

func (p *DatacenterState) NodeDatacenter() *datacenter.Datacenter {
	return p.nodeDatacenter
}

func (p *DatacenterState) Refresh(pkg *Package,
	db *database.Database) (err error) {

	dcId := node.Self.Datacenter

	err = datacenterCache.Sync(db)
	if err != nil {
		return
	}

	if dcId.IsZero() {
		p.nodeDatacenter = nil
		pkg.Evict()
		return
	}

	dcs, err := Decode[datacenter.Datacenter](datacenterCache)
	if err != nil {
		return
	}

	if len(dcs) == 0 {
		err = &database.NotFoundError{
			errors.New("statec: Datacenter not found"),
		}
		return
	}

	p.nodeDatacenter = dcs[0]
	pkg.Cache(15 * time.Second)

	return
}

func (p *DatacenterState) Apply(st *State) {
	st.NodeDatacenter = p.NodeDatacenter
}

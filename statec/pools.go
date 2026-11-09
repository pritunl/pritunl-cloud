package statec

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/pool"
)

var (
	Pools      = &PoolsState{}
	PoolsPkg   = NewPackage(Pools)
	poolsCache = NewCache(
		"pools",
		[]string{"zone"},
		func() Scope {
			zneId := node.Self.Zone
			if zneId.IsZero() {
				return Scope{}
			}

			return Scope{
				Filter: bson.M{
					"zone": zneId,
				},
				Key: zneId.Hex(),
				Ok:  true,
			}
		},
		func(evt *Event) bool {
			zneId, ok := evt.DocObjectId("zone")
			return ok && zneId == node.Self.Zone
		},
	)
)

type PoolsState struct {
	nodePools []*pool.Pool
}

func (p *PoolsState) NodePools() []*pool.Pool {
	return p.nodePools
}

func (p *PoolsState) Refresh(pkg *Package,
	db *database.Database) (err error) {

	zneId := node.Self.Zone

	err = poolsCache.Sync(db)
	if err != nil {
		return
	}

	if zneId.IsZero() {
		p.nodePools = nil
		return
	}

	pools, err := Decode[pool.Pool](poolsCache)
	if err != nil {
		return
	}
	p.nodePools = pools

	return
}

func (p *PoolsState) Apply(st *State) {
	st.NodePools = p.NodePools
}

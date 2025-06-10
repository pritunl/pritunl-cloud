package state

import (
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/pool"
)

var (
	Pools    = &PoolsState{}
	PoolsPkg = NewPackage(Pools)
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
	if zneId.IsZero() {
		p.nodePools = nil
		return
	}

	pools, err := pool.GetAll(db, &bson.M{
		"zone": zneId,
	})
	if err != nil {
		return
	}
	p.nodePools = pools

	return
}

func (p *PoolsState) Apply(st *State) {
	st.NodePools = p.NodePools
}

package state

import (
	"time"

	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/datacenter"
	"github.com/pritunl/pritunl-cloud/node"
)

var (
	Datacenter    = &DatacenterState{}
	DatacenterPkg = NewPackage(Datacenter)
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
	if dcId.IsZero() {
		p.nodeDatacenter = nil
		pkg.Evict()
		return
	}

	dc, e := datacenter.Get(db, dcId)
	if e != nil {
		err = e
		return
	}

	p.nodeDatacenter = dc
	pkg.Cache(15 * time.Second)

	return
}

func (p *DatacenterState) Apply(st *State) {
	st.NodeDatacenter = p.NodeDatacenter
}

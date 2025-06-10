package state

import (
	"time"

	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/zone"
)

var (
	Zone    = &ZoneState{}
	ZonePkg = NewPackage(Zone)
)

type ZoneState struct {
	nodeZone *zone.Zone
}

func (p *ZoneState) NodeZone() *zone.Zone {
	return p.nodeZone
}

func (p *ZoneState) Refresh(pkg *Package,
	db *database.Database) (err error) {

	zneId := node.Self.Zone
	if zneId.IsZero() {
		p.nodeZone = nil
		pkg.Evict()
		return
	}

	zne, e := zone.Get(db, zneId)
	if e != nil {
		err = e
		return
	}

	p.nodeZone = zne

	pkg.Cache(15 * time.Second)

	return
}

func (p *ZoneState) Apply(st *State) {
	st.NodeZone = p.NodeZone
}

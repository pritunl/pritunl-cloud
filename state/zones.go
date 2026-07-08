package state

import (
	"time"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/zone"
)

var (
	Zones    = &ZonesState{}
	ZonesPkg = NewPackage(Zones)
)

type ZonesState struct {
	vxlan   bool
	zoneMap map[bson.ObjectID]*zone.Zone
	nodes   []*node.Node
}

func (p *ZonesState) VxLan() bool {
	return p.vxlan
}

func (p *ZonesState) GetZone(zneId bson.ObjectID) *zone.Zone {
	return p.zoneMap[zneId]
}

func (p *ZonesState) Nodes() []*node.Node {
	return p.nodes
}

func (p *ZonesState) Refresh(pkg *Package,
	db *database.Database) (err error) {

	nodeDc := Datacenter.NodeDatacenter()
	if nodeDc == nil {
		p.vxlan = false
		p.zoneMap = nil
		p.nodes = nil
		pkg.Evict()
		return
	}

	znes, e := zone.GetAllDatacenter(db, nodeDc.Id)
	if e != nil {
		err = e
		return
	}

	zonesMap := map[bson.ObjectID]*zone.Zone{}
	for _, zne := range znes {
		zonesMap[zne.Id] = zne
	}
	p.zoneMap = zonesMap

	if nodeDc.Vxlan() {
		p.vxlan = true

		ndes, e := node.GetAllNet(db)
		if e != nil {
			err = e
			return
		}

		p.nodes = ndes
	} else {
		p.vxlan = false
	}

	pkg.Cache(10 * time.Second)

	return
}

func (p *ZonesState) Apply(st *State) {
	st.VxLan = p.VxLan
	st.GetZone = p.GetZone
	st.Nodes = p.Nodes
}

func init() {
	ZonesPkg.
		After(Datacenter)
}

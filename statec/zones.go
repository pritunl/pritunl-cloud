package statec

import (
	"bytes"
	"sort"
	"time"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/zone"
)

var (
	Zones      = &ZonesState{}
	ZonesPkg   = NewPackage(Zones)
	zonesCache = NewCache(
		"zones",
		[]string{"datacenter"},
		func() Scope {
			dcId := node.Self.Datacenter
			if dcId.IsZero() {
				return Scope{}
			}

			return Scope{
				Filter: bson.M{
					"datacenter": dcId,
				},
				Key: dcId.Hex(),
				Ok:  true,
			}
		},
		func(evt *Event) bool {
			dcId, ok := evt.DocObjectId("datacenter")
			return ok && dcId == node.Self.Datacenter
		},
	)
	nodesCache = NewCache(
		"nodes",
		nil,
		func() Scope {
			nodeDc := Datacenter.NodeDatacenter()
			if nodeDc == nil || !nodeDc.Vxlan() {
				return Scope{}
			}

			return Scope{
				Filter: bson.M{},
				Key:    "vxlan",
				Ok:     true,
			}
		},
		nil,
	).Only("datacenter", "zone", "private_ips").SetProjection(&bson.D{
		{"datacenter", 1},
		{"zone", 1},
		{"private_ips", 1},
	})
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

	err = zonesCache.Sync(db)
	if err != nil {
		return
	}

	err = nodesCache.Sync(db)
	if err != nil {
		return
	}

	if nodeDc == nil {
		p.vxlan = false
		p.zoneMap = nil
		p.nodes = nil
		pkg.Evict()
		return
	}

	znes, err := Decode[zone.Zone](zonesCache)
	if err != nil {
		return
	}

	sort.SliceStable(znes, func(i, j int) bool {
		if znes[i].Name != znes[j].Name {
			return znes[i].Name < znes[j].Name
		}
		return bytes.Compare(znes[i].Id[:], znes[j].Id[:]) < 0
	})

	zonesMap := map[bson.ObjectID]*zone.Zone{}
	for _, zne := range znes {
		zonesMap[zne.Id] = zne
	}
	p.zoneMap = zonesMap

	if nodeDc.Vxlan() {
		p.vxlan = true

		ndes, e := Decode[node.Node](nodesCache)
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

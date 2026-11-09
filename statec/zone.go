package statec

import (
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/zone"
)

var (
	Zone      = &ZoneState{}
	ZonePkg   = NewPackage(Zone)
	zoneCache = NewCache(
		"zones",
		nil,
		func() Scope {
			zneId := node.Self.Zone
			if zneId.IsZero() {
				return Scope{}
			}

			return Scope{
				Filter: bson.M{
					"_id": zneId,
				},
				Key: zneId.Hex(),
				Ok:  true,
			}
		},
		func(evt *Event) bool {
			zneId, ok := evt.DocId()
			return ok && zneId == node.Self.Zone
		},
	)
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

	err = zoneCache.Sync(db)
	if err != nil {
		return
	}

	if zneId.IsZero() {
		p.nodeZone = nil
		pkg.Evict()
		return
	}

	znes, err := Decode[zone.Zone](zoneCache)
	if err != nil {
		return
	}

	if len(znes) == 0 {
		err = &database.NotFoundError{
			errors.New("statec: Zone not found"),
		}
		return
	}

	p.nodeZone = znes[0]

	pkg.Cache(15 * time.Second)

	return
}

func (p *ZoneState) Apply(st *State) {
	st.NodeZone = p.NodeZone
}

package statec

import (
	"sort"
	"strings"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/node"
)

var (
	Disks      = &DisksState{}
	DisksPkg   = NewPackage(Disks)
	disksCache = NewCache(
		"disks",
		[]string{"node", "pool"},
		func() Scope {
			ndeId := node.Self.Id
			ndePools := node.Self.Pools
			if ndePools == nil {
				ndePools = []bson.ObjectID{}
			}

			return Scope{
				Filter: bson.M{
					"$or": []bson.M{
						{"node": ndeId},
						{"pool": &bson.M{
							"$in": ndePools,
						}},
					},
				},
				Key: ndeId.Hex() + ":" + idsKey(ndePools),
				Ok:  true,
			}
		},
		func(evt *Event) bool {
			ndeId, ok := evt.DocObjectId("node")
			if ok && ndeId == node.Self.Id {
				return true
			}

			poolId, ok := evt.DocObjectId("pool")
			if ok {
				for _, ndePool := range node.Self.Pools {
					if ndePool == poolId {
						return true
					}
				}
			}

			return false
		},
	)
)

func idsKey(ids []bson.ObjectID) string {
	hexes := make([]string, 0, len(ids))
	for _, id := range ids {
		hexes = append(hexes, id.Hex())
	}
	sort.Strings(hexes)
	return strings.Join(hexes, ",")
}

type DisksState struct {
	disks           []*disk.Disk
	instanceDisks   map[bson.ObjectID][]*disk.Disk
	deploymentDisks map[bson.ObjectID][]*disk.Disk
}

func (p *DisksState) Disks() []*disk.Disk {
	return p.disks
}

func (p *DisksState) GetInstaceDisks(instId bson.ObjectID) []*disk.Disk {
	return p.instanceDisks[instId]
}

func (p *DisksState) GetDeploymentDisks(
	deplyId bson.ObjectID) []*disk.Disk {

	return p.deploymentDisks[deplyId]
}

func (p *DisksState) InstaceDisksMap() map[bson.ObjectID][]*disk.Disk {
	return p.instanceDisks
}

func (p *DisksState) Refresh(pkg *Package,
	db *database.Database) (err error) {

	err = disksCache.Sync(db)
	if err != nil {
		return
	}

	disks, err := Decode[disk.Disk](disksCache)
	if err != nil {
		return
	}
	p.disks = disks

	instanceDisks := map[bson.ObjectID][]*disk.Disk{}
	deploymentDisks := map[bson.ObjectID][]*disk.Disk{}
	for _, dsk := range disks {
		if !dsk.Instance.IsZero() {
			instanceDisks[dsk.Instance] = append(
				instanceDisks[dsk.Instance], dsk)
		}
		if !dsk.Deployment.IsZero() {
			deploymentDisks[dsk.Deployment] = append(
				deploymentDisks[dsk.Deployment], dsk)
		}
	}
	p.instanceDisks = instanceDisks
	p.deploymentDisks = deploymentDisks

	return
}

func (p *DisksState) Apply(st *State) {
	st.Disks = p.Disks
	st.GetInstaceDisks = p.GetInstaceDisks
	st.GetDeploymentDisks = p.GetDeploymentDisks
	st.InstaceDisksMap = p.InstaceDisksMap
}

package statec

import (
	"sync"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/vpc"
)

var (
	Vpcs      = &VpcsState{}
	VpcsPkg   = NewPackage(Vpcs)
	vpcsCache = NewCache(
		"vpcs",
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
	vpcIpsCache = NewCache(
		"vpcs_ip",
		[]string{"vpc"},
		func() Scope {
			if node.Self.Datacenter.IsZero() {
				return Scope{}
			}

			ids := Vpcs.vpcIds()
			return Scope{
				Filter: bson.M{
					"vpc": bson.M{
						"$in": ids,
					},
				},
				Key: idsKey(ids),
				Ok:  true,
			}
		},
		func(evt *Event) bool {
			vpcId, ok := evt.DocObjectId("vpc")
			return ok && Vpcs.hasVpc(vpcId)
		},
	)
)

type VpcsState struct {
	vpcs      []*vpc.Vpc
	vpcsMap   map[bson.ObjectID]*vpc.Vpc
	vpcIpsMap map[bson.ObjectID][]*vpc.VpcIp

	idsLock sync.Mutex
	ids     []bson.ObjectID
	idsSet  map[bson.ObjectID]struct{}
}

func (p *VpcsState) vpcIds() (ids []bson.ObjectID) {
	p.idsLock.Lock()
	ids = make([]bson.ObjectID, len(p.ids))
	copy(ids, p.ids)
	p.idsLock.Unlock()
	return
}

func (p *VpcsState) hasVpc(vpcId bson.ObjectID) (ok bool) {
	p.idsLock.Lock()
	_, ok = p.idsSet[vpcId]
	p.idsLock.Unlock()
	return
}

func (p *VpcsState) setVpcIds(ids []bson.ObjectID) {
	idsSet := map[bson.ObjectID]struct{}{}
	for _, id := range ids {
		idsSet[id] = struct{}{}
	}

	p.idsLock.Lock()
	p.ids = ids
	p.idsSet = idsSet
	p.idsLock.Unlock()
}

func (p *VpcsState) Vpc(vpcId bson.ObjectID) *vpc.Vpc {
	return p.vpcsMap[vpcId]
}

func (p *VpcsState) VpcsMap() map[bson.ObjectID]*vpc.Vpc {
	return p.vpcsMap
}

func (p *VpcsState) VpcIps(vpcId bson.ObjectID) []*vpc.VpcIp {
	return p.vpcIpsMap[vpcId]
}

func (p *VpcsState) VpcIpsMap() map[bson.ObjectID][]*vpc.VpcIp {
	return p.vpcIpsMap
}

func (p *VpcsState) Vpcs() []*vpc.Vpc {
	return p.vpcs
}

func (p *VpcsState) Refresh(pkg *Package,
	db *database.Database) (err error) {

	dcId := node.Self.Datacenter
	vpcsId := []bson.ObjectID{}
	vpcsMap := map[bson.ObjectID]*vpc.Vpc{}
	if dcId.IsZero() {
		p.vpcs = nil
		p.vpcsMap = map[bson.ObjectID]*vpc.Vpc{}
		p.vpcIpsMap = map[bson.ObjectID][]*vpc.VpcIp{}
		p.setVpcIds(vpcsId)

		err = vpcsCache.Sync(db)
		if err != nil {
			return
		}

		err = vpcIpsCache.Sync(db)
		if err != nil {
			return
		}
		return
	}

	err = vpcsCache.Sync(db)
	if err != nil {
		return
	}

	vpcs, err := Decode[vpc.Vpc](vpcsCache)
	if err != nil {
		return
	}

	for _, vc := range vpcs {
		vpcsId = append(vpcsId, vc.Id)
		vpcsMap[vc.Id] = vc
	}

	p.vpcs = vpcs
	p.vpcsMap = vpcsMap
	p.setVpcIds(vpcsId)

	err = vpcIpsCache.Sync(db)
	if err != nil {
		return
	}

	vpcIps, err := Decode[vpc.VpcIp](vpcIpsCache)
	if err != nil {
		return
	}

	vpcIpsMap := map[bson.ObjectID][]*vpc.VpcIp{}
	for _, vpcIp := range vpcIps {
		vpcIpsMap[vpcIp.Vpc] = append(vpcIpsMap[vpcIp.Vpc], vpcIp)
	}
	p.vpcIpsMap = vpcIpsMap

	return
}

func (p *VpcsState) Apply(st *State) {
	st.Vpc = p.Vpc
	st.VpcsMap = p.VpcsMap
	st.VpcIps = p.VpcIps
	st.VpcIpsMap = p.VpcIpsMap
	st.Vpcs = p.Vpcs
}

package state

import (
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/vpc"
)

var (
	Vpcs    = &VpcsState{}
	VpcsPkg = NewPackage(Vpcs)
)

type VpcsState struct {
	vpcs      []*vpc.Vpc
	vpcsMap   map[primitive.ObjectID]*vpc.Vpc
	vpcIpsMap map[primitive.ObjectID][]*vpc.VpcIp
}

func (p *VpcsState) Vpc(vpcId primitive.ObjectID) *vpc.Vpc {
	return p.vpcsMap[vpcId]
}

func (p *VpcsState) VpcsMap() map[primitive.ObjectID]*vpc.Vpc {
	return p.vpcsMap
}

func (p *VpcsState) VpcIps(vpcId primitive.ObjectID) []*vpc.VpcIp {
	return p.vpcIpsMap[vpcId]
}

func (p *VpcsState) VpcIpsMap() map[primitive.ObjectID][]*vpc.VpcIp {
	return p.vpcIpsMap
}

func (p *VpcsState) Vpcs() []*vpc.Vpc {
	return p.vpcs
}

func (p *VpcsState) Refresh(pkg *Package,
	db *database.Database) (err error) {

	dcId := node.Self.Datacenter
	vpcsId := []primitive.ObjectID{}
	vpcsMap := map[primitive.ObjectID]*vpc.Vpc{}
	if dcId.IsZero() {
		p.vpcs = nil
		p.vpcsMap = map[primitive.ObjectID]*vpc.Vpc{}
		p.vpcIpsMap = map[primitive.ObjectID][]*vpc.VpcIp{}
		return
	}

	vpcs, err := vpc.GetDatacenter(db, dcId)
	if err != nil {
		return
	}

	for _, vc := range vpcs {
		vpcsId = append(vpcsId, vc.Id)
		vpcsMap[vc.Id] = vc
	}

	p.vpcs = vpcs
	p.vpcsMap = vpcsMap

	vpcIpsMap, err := vpc.GetIpsMapped(db, vpcsId)
	if err != nil {
		return
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

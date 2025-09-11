package state

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/node"
)

var (
	Disks    = &DisksState{}
	DisksPkg = NewPackage(Disks)
)

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

	ndeId := node.Self.Id
	ndePools := node.Self.Pools

	disks, err := disk.GetNode(db, ndeId, ndePools)
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

package state

import (
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/node"
)

var (
	Disks    = &DisksState{}
	DisksPkg = NewPackage(Disks)
)

type DisksState struct {
	disks         []*disk.Disk
	instanceDisks map[primitive.ObjectID][]*disk.Disk
}

func (p *DisksState) Disks() []*disk.Disk {
	return p.disks
}

func (p *DisksState) GetInstaceDisks(instId primitive.ObjectID) []*disk.Disk {
	return p.instanceDisks[instId]
}

func (p *DisksState) InstaceDisksMap() map[primitive.ObjectID][]*disk.Disk {
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

	instanceDisks := map[primitive.ObjectID][]*disk.Disk{}
	for _, dsk := range disks {
		dsks := instanceDisks[dsk.Instance]
		if dsks == nil {
			dsks = []*disk.Disk{}
		}
		instanceDisks[dsk.Instance] = append(dsks, dsk)
	}
	p.instanceDisks = instanceDisks

	return
}

func (p *DisksState) Apply(st *State) {
	st.Disks = p.Disks
	st.GetInstaceDisks = p.GetInstaceDisks
	st.InstaceDisksMap = p.InstaceDisksMap
}

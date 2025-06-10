package state

import (
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/qemu"
	"github.com/pritunl/pritunl-cloud/vm"
)

var (
	Virtuals    = &VirtualsState{}
	VirtualsPkg = NewPackage(Virtuals)
)

type VirtualsState struct {
	virtsMap map[primitive.ObjectID]*vm.VirtualMachine
}

func (p *VirtualsState) DiskInUse(instId, dskId primitive.ObjectID) bool {
	curVirt := p.virtsMap[instId]

	if curVirt != nil {
		if curVirt.State != vm.Stopped && curVirt.State != vm.Failed {
			for _, vmDsk := range curVirt.Disks {
				if vmDsk.GetId() == dskId {
					return true
				}
			}
		}
	}

	return false
}

func (p *VirtualsState) GetVirt(instId primitive.ObjectID) *vm.VirtualMachine {
	if instId.IsZero() {
		return nil
	}
	return p.virtsMap[instId]
}

func (p *VirtualsState) VirtsMap() map[primitive.ObjectID]*vm.VirtualMachine {
	return p.virtsMap
}

func (p *VirtualsState) Refresh(pkg *Package,
	db *database.Database) (err error) {

	curVirts, err := qemu.GetVms(db)
	if err != nil {
		return
	}

	virtsMap := map[primitive.ObjectID]*vm.VirtualMachine{}
	for _, virt := range curVirts {
		virtsMap[virt.Id] = virt
	}
	p.virtsMap = virtsMap

	return
}

func (p *VirtualsState) Apply(st *State) {
	st.DiskInUse = p.DiskInUse
	st.GetVirt = p.GetVirt
	st.VirtsMap = p.VirtsMap
}

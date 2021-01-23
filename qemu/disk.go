package qemu

import (
	"time"

	"github.com/pritunl/pritunl-cloud/qms"
	"github.com/pritunl/pritunl-cloud/store"
	"github.com/pritunl/pritunl-cloud/vm"
)

func UpdateVmDisk(virt *vm.VirtualMachine) (err error) {
	for i := 0; i < 20; i++ {
		if virt.State == vm.Running {
			disks, e := qms.GetDisks(virt.Id)
			if e != nil {
				if i < 19 {
					time.Sleep(100 * time.Millisecond)
					_ = UpdateVmState(virt)
					continue
				}
				err = e

				return
			}
			virt.Disks = disks

			store.SetDisks(virt.Id, disks)
		}

		break
	}

	return
}

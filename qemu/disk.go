package qemu

import (
	"time"

	"github.com/pritunl/pritunl-cloud/qmp"
	"github.com/pritunl/pritunl-cloud/store"
	"github.com/pritunl/pritunl-cloud/vm"
)

func UpdateVmDisk(virt *vm.VirtualMachine) (err error) {
	for i := 0; i < 10; i++ {
		if virt.State == vm.Running {
			_, disks, e := qmp.GetDisks(virt.Id)
			if e != nil {
				if i < 9 {
					time.Sleep(300 * time.Millisecond)
					_ = UpdateState(virt)
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

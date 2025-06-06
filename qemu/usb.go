package qemu

import (
	"time"

	"github.com/pritunl/pritunl-cloud/qms"
	"github.com/pritunl/pritunl-cloud/store"
	"github.com/pritunl/pritunl-cloud/vm"
)

func UpdateVmUsb(virt *vm.VirtualMachine) (err error) {
	for i := 0; i < 10; i++ {
		if virt.State == vm.Running {
			usbs, e := qms.GetUsbDevices(virt.Id)
			if e != nil {
				if i < 9 {
					time.Sleep(300 * time.Millisecond)
					_ = UpdateState(virt)
					continue
				}
				err = e

				return
			}
			virt.UsbDevices = usbs

			store.SetUsbs(virt.Id, usbs)
		}

		break
	}

	return
}

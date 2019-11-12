package qemu

import (
	"encoding/json"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/vm"
)

func NewQemu(virt *vm.VirtualMachine) (qm *Qemu, err error) {
	data, err := json.Marshal(virt)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "qemu: Failed to marshal virt"),
		}
		return
	}

	qm = &Qemu{
		Id:         virt.Id,
		Data:       string(data),
		Kvm:        node.Self.Hypervisor == node.Kvm,
		Machine:    "pc",
		Cpu:        "host",
		Cpus:       virt.Processors,
		Cores:      1,
		Threads:    1,
		Boot:       "c",
		Memory:     virt.Memory,
		Vnc:        virt.Vnc,
		VncDisplay: virt.VncDisplay,
		Disks:      []*Disk{},
		Networks:   []*Network{},
		UsbDevices: []*UsbDevice{},
	}

	for _, disk := range virt.Disks {
		qm.Disks = append(qm.Disks, &Disk{
			Media:   "disk",
			Index:   disk.Index,
			File:    disk.Path,
			Format:  "qcow2",
			Discard: false,
		})
	}

	for i, net := range virt.NetworkAdapters {
		qm.Networks = append(qm.Networks, &Network{
			MacAddress: net.MacAddress,
			Iface:      vm.GetIface(virt.Id, i),
		})
	}

	for _, device := range virt.UsbDevices {
		qm.UsbDevices = append(qm.UsbDevices, &UsbDevice{
			Vendor:  device.Vendor,
			Product: device.Product,
		})
	}

	return
}

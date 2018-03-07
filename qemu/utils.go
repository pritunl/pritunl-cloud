package qemu

import (
	"encoding/json"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
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
		Id:       virt.Id,
		Data:     string(data),
		Kvm:      true,
		Machine:  "pc",
		Accel:    "kvm",
		Cpu:      "host",
		Cpus:     1,
		Cores:    virt.Processors,
		Threads:  1,
		Boot:     "c",
		Memory:   virt.Memory,
		Disks:    []*Disk{},
		Networks: []*Network{},
	}

	for _, disk := range virt.Disks {
		qm.Disks = append(qm.Disks, &Disk{
			Media:   "disk",
			Index:   disk.Index,
			File:    disk.Path,
			Format:  "qcow2",
			Discard: true,
		})
	}

	for i, net := range virt.NetworkAdapters {
		switch net.Type {
		case vm.Bridge:
			qm.Networks = append(qm.Networks, &Network{
				Type:       "nic",
				MacAddress: net.MacAddress,
			})
			qm.Networks = append(qm.Networks, &Network{
				Type:   "bridge",
				Iface:  vm.GetIface(virt.Id, i),
				Bridge: net.HostInterface,
			})
			break
		case vm.Vxlan:
			qm.Networks = append(qm.Networks, &Network{
				Type:       "nic",
				MacAddress: net.MacAddress,
			})
			qm.Networks = append(qm.Networks, &Network{
				Type:   "bridge",
				Iface:  vm.GetIface(virt.Id, i),
				Bridge: net.HostInterface,
			})
			break
		default:
			err = &errortypes.ParseError{
				errors.Newf("qemu: Unknown network adapter type %s",
					net.Type),
			}
			return
		}
	}

	return
}

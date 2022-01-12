package qemu

import (
	"encoding/json"
	"sort"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/paths"
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

	ovmfCodePath := ""
	if virt.Uefi {
		ovmfCodePath, err = paths.FindOvmfCodePath(virt.SecureBoot)
		if err != nil {
			return
		}
	}

	guiUser := node.Self.GuiUser
	guiMode := node.Self.GuiMode
	if guiMode == "" {
		guiMode = node.Sdl
	}

	qm = &Qemu{
		Id:           virt.Id,
		Data:         string(data),
		Kvm:          node.Self.Hypervisor == node.Kvm,
		Machine:      "q35",
		Cpu:          "host",
		Cores:        virt.Processors,
		Threads:      1,
		Dies:         1,
		Sockets:      1,
		Boot:         "c",
		Uefi:         virt.Uefi,
		SecureBoot:   virt.SecureBoot,
		OvmfCodePath: ovmfCodePath,
		OvmfVarsPath: paths.GetOvmfVarsPath(virt.Id),
		Memory:       virt.Memory,
		Hugepages:    virt.Hugepages,
		Vnc:          virt.Vnc && virt.VncDisplay != 0,
		VncDisplay:   virt.VncDisplay,
		Spice:        virt.Spice && virt.SpicePort != 0,
		SpicePort:    virt.SpicePort,
		Gui:          virt.Gui && node.Self.Gui && guiUser != "",
		GuiUser:      guiUser,
		GuiMode:      guiMode,
		Disks:        []*Disk{},
		Networks:     []*Network{},
		Isos:         []*Iso{},
		UsbDevices:   []*UsbDevice{},
		PciDevices:   []*PciDevice{},
		DriveDevices: []*DriveDevice{},
		IscsiDevices: []*IscsiDevice{},
	}

	for _, disk := range virt.Disks {
		qm.Disks = append(qm.Disks, &Disk{
			Id:     disk.Id.Hex(),
			Index:  disk.Index,
			File:   disk.Path,
			Format: "qcow2",
		})
	}

	sort.Sort(qm.Disks)

	for i, net := range virt.NetworkAdapters {
		qm.Networks = append(qm.Networks, &Network{
			MacAddress: net.MacAddress,
			Iface:      vm.GetIface(virt.Id, i),
		})
	}

	for _, is := range virt.Isos {
		qm.Isos = append(qm.Isos, &Iso{
			Name: is.Name,
		})
	}

	for _, device := range virt.UsbDevices {
		qm.UsbDevices = append(qm.UsbDevices, &UsbDevice{
			Vendor:  device.Vendor,
			Product: device.Product,
			Bus:     device.Bus,
			Address: device.Address,
		})
	}

	for _, device := range virt.PciDevices {
		qm.PciDevices = append(qm.PciDevices, &PciDevice{
			Slot: device.Slot,
		})
	}

	for _, device := range virt.DriveDevices {
		qm.DriveDevices = append(qm.DriveDevices, &DriveDevice{
			Id: device.Id,
		})
	}

	for _, device := range virt.IscsiDevices {
		qm.IscsiDevices = append(qm.IscsiDevices, &IscsiDevice{
			Uri: device.Uri,
		})
	}

	return
}

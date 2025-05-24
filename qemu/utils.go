package qemu

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/pci"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/sirupsen/logrus"
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

	namespace := ""
	if !virt.HasExternalNetwork() {
		namespace = vm.GetNamespace(virt.Id, 0)
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
		Tpm:          virt.Tpm,
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
		ProtectHome:  virt.ProtectHome(),
		ProtectTmp:   virt.ProtectTmp(),
		Namespace:    namespace,
		Disks:        []*Disk{},
		Networks:     []*Network{},
		Isos:         []*Iso{},
		UsbDevices:   []*UsbDevice{},
		PciDevices:   []*PciDevice{},
		DriveDevices: []*DriveDevice{},
		IscsiDevices: []*IscsiDevice{},
		Mounts:       []*Mount{},
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
		dev, e := pci.GetVfio(device.Slot)
		if e != nil {
			err = e
			return
		}

		if dev == nil {
			logrus.WithFields(logrus.Fields{
				"instance_id": virt.Id.Hex(),
				"device_slot": device.Slot,
			}).Error("qemu: Failed to find vfio device")

			continue
		}

		name := strings.ToLower(dev.Name)

		qm.PciDevices = append(qm.PciDevices, &PciDevice{
			Slot: device.Slot,
			Gpu: strings.Contains(name, "vga compatible") ||
				strings.Contains(name, "vga controller") ||
				strings.Contains(name, "graphics controller") ||
				strings.Contains(name, "display controller"),
		})
	}

	for _, device := range virt.DriveDevices {
		qm.DriveDevices = append(qm.DriveDevices, &DriveDevice{
			Id:     device.Id,
			Type:   device.Type,
			VgName: device.VgName,
			LvName: device.LvName,
		})
	}

	for _, device := range virt.IscsiDevices {
		qm.IscsiDevices = append(qm.IscsiDevices, &IscsiDevice{
			Uri: device.Uri,
		})
	}

	for _, mount := range virt.Mounts {
		shareId := paths.GetShareId(virt.Id, mount.Name)
		sockPath := paths.GetShareSockPath(virt.Id, shareId)

		qm.Mounts = append(qm.Mounts, &Mount{
			Id:   shareId,
			Name: utils.FilterNameCmd(mount.Name),
			Sock: sockPath,
		})
	}

	return
}

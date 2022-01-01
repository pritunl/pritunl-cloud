package qemu

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
)

func GetQemuVersion() (major, minor, patch int, err error) {
	qemuPath, err := GetQemuPath()
	if err != nil {
		return
	}

	output, _ := utils.ExecCombinedOutputLogged(
		nil,
		qemuPath, "--version",
	)

	lines := strings.Split(output, "\n")

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 4 || strings.ToLower(fields[2]) != "version" {
			continue
		}

		versions := strings.Split(fields[3], ".")
		if len(versions) != 3 {
			continue
		}

		var e error
		major, e = strconv.Atoi(versions[0])
		if e != nil {
			continue
		}

		minor, e = strconv.Atoi(versions[1])
		if e != nil {
			major = 0
			continue
		}

		patch, e = strconv.Atoi(versions[1])
		if e != nil {
			major = 0
			minor = 0
			continue
		}

		break
	}

	if major == 0 {
		err = &errortypes.ParseError{
			errors.Newf("qemu: Invalid Qemu version '%s'", output),
		}
		return
	}

	return
}

func GetQemuPath() (path string, err error) {
	exists, err := utils.Exists(System)
	if err != nil {
		return
	}
	if exists {
		path = System
	} else {
		path = Libvirt
	}

	return
}

func GetUringSupport() (supported bool, err error) {
	major, minor, _, err := GetQemuVersion()
	if err != nil {
		return
	}

	if major > 6 {
		supported = true
	} else if major == 6 && minor >= 2 {
		supported = true
	}

	return
}

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
		Vnc:          virt.Vnc,
		VncDisplay:   virt.VncDisplay,
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

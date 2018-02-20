package qemu

import (
	"encoding/json"
	"fmt"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/vm"
	"gopkg.in/mgo.v2/bson"
	"path"
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
			File:    disk.Path,
			Format:  "qcow2",
			Discard: true,
		})
	}

	for i, net := range virt.NetworkAdapters {
		qm.Networks = append(qm.Networks, &Network{
			Type:       "nic",
			MacAddress: net.MacAddress,
		})
		qm.Networks = append(qm.Networks, &Network{
			Type:   "bridge",
			Iface:  vm.GetIface(virt.Id, i),
			Bridge: net.HostInterface,
		})
	}

	return
}

func GetUnitName(virtId bson.ObjectId) string {
	return fmt.Sprintf("pritunl_cloud_%s.service", virtId.Hex())
}

func GetUnitPath(virtId bson.ObjectId) string {
	return path.Join(settings.Hypervisor.SystemdPath, GetUnitName(virtId))
}

func GetPidPath(virtId bson.ObjectId) string {
	return path.Join(settings.Hypervisor.LibPath,
		fmt.Sprintf("%s.pid", virtId.Hex()))
}

func GetSockPath(virtId bson.ObjectId) string {
	return path.Join(settings.Hypervisor.LibPath,
		fmt.Sprintf("%s.sock", virtId.Hex()))
}

func GetGuestPath(virtId bson.ObjectId) string {
	return path.Join(settings.Hypervisor.LibPath,
		fmt.Sprintf("%s.guest", virtId.Hex()))
}

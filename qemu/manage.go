package qemu

import (
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/cloudinit"
	"github.com/pritunl/pritunl-cloud/data"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/qga"
	"github.com/pritunl/pritunl-cloud/qms"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/systemd"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"net"
	"regexp"
	"strings"
	"sync"
	"time"
)

var (
	serviceReg = regexp.MustCompile("pritunl_cloud_([a-z0-9]+).service")
)

func GetVmInfo(vmId bson.ObjectId, getDisks bool) (
	virt *vm.VirtualMachine, err error) {

	unitPath := paths.GetUnitPath(vmId)

	unitData, err := ioutil.ReadFile(unitPath)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "qemu: Failed to read service"),
		}
		return
	}

	virt = &vm.VirtualMachine{}
	for _, line := range strings.Split(string(unitData), "\n") {
		if !strings.HasPrefix(line, "PritunlData=") {
			continue
		}

		lineSpl := strings.SplitN(line, "=", 2)
		if len(lineSpl) != 2 || len(lineSpl[1]) < 6 {
			continue
		}

		err = json.Unmarshal([]byte(lineSpl[1]), virt)
		if err != nil {
			err = &errortypes.ParseError{
				errors.Wrap(err, "qemu: Failed to parse service data"),
			}
			return
		}

		break
	}

	if virt.Id == "" {
		virt = nil
		return
	}

	unitName := paths.GetUnitName(virt.Id)
	state, err := systemd.GetState(unitName)
	if err != nil {
		return
	}

	switch state {
	case "active":
		virt.State = vm.Running
		break
	case "deactivating":
		virt.State = vm.Stopped
		break
	case "inactive":
		virt.State = vm.Stopped
		break
	case "failed":
		virt.State = vm.Failed
		break
	case "unknown":
		virt.State = vm.Stopped
		break
	default:
		logrus.WithFields(logrus.Fields{
			"id":    virt.Id.Hex(),
			"state": state,
		}).Info("qemu: Unknown virtual machine state")
		virt.State = vm.Failed
		break
	}

	if virt.State == vm.Running && getDisks {
		disks, e := qms.GetDisks(vmId)
		if e != nil {
			err = e
			return
		}
		virt.Disks = disks
	}

	guestPath := paths.GetGuestPath(virt.Id)
	ifaces, err := qga.GetInterfaces(guestPath)
	if err == nil {
		for _, adapter := range virt.NetworkAdapters {
			ipAddr, ipAddr6 := ifaces.GetAddr(adapter.MacAddress)
			if ipAddr != "" {
				adapter.IpAddress = ipAddr
				adapter.IpAddress6 = ipAddr6
			}
		}
	} else {
		err = nil
	}

	return
}

func GetVms(db *database.Database) (virts []*vm.VirtualMachine, err error) {
	systemdPath := settings.Hypervisor.SystemdPath
	virts = []*vm.VirtualMachine{}

	items, err := ioutil.ReadDir(systemdPath)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "qemu: Failed to read systemd directory"),
		}
		return
	}

	units := []string{}
	for _, item := range items {
		if strings.HasPrefix(item.Name(), "pritunl_cloud") {
			units = append(units, item.Name())
		}
	}

	waiter := sync.WaitGroup{}
	virtsLock := sync.Mutex{}

	for _, unit := range units {
		match := serviceReg.FindStringSubmatch(unit)
		if match == nil || len(match) != 2 || !bson.IsObjectIdHex(match[1]) {
			continue
		}
		vmId := bson.ObjectIdHex(match[1])

		waiter.Add(1)
		go func() {
			defer waiter.Done()

			virt, e := GetVmInfo(vmId, true)
			if e != nil {
				err = e
				return
			}

			if virt != nil {
				e = virt.Commit(db)
				if e != nil {
					logrus.WithFields(logrus.Fields{
						"error": e,
					}).Error("qemu: Failed to commit VM state")
				}

				virtsLock.Lock()
				virts = append(virts, virt)
				virtsLock.Unlock()
			}
		}()
	}

	if err != nil {
		return
	}

	waiter.Wait()

	return
}

func Wait(db *database.Database, virt *vm.VirtualMachine) (err error) {
	unitName := paths.GetUnitName(virt.Id)

	var vrt *vm.VirtualMachine
	for i := 0; i < settings.Hypervisor.StartTimeout; i++ {
		vrt, err = GetVmInfo(virt.Id, false)
		if err != nil {
			return
		}

		if vrt.State == vm.Running {
			break
		}
	}

	if vrt == nil || vrt.State != vm.Running {
		err = systemd.Stop(unitName)
		if err != nil {
			return
		}

		err = &errortypes.TimeoutError{
			errors.New("qemu: Power on timeout"),
		}

		return
	}

	return
}

func NetworkConf(db *database.Database, virt *vm.VirtualMachine) (err error) {
	ifaceNames := set.NewSet()
	for i := range virt.NetworkAdapters {
		ifaceNames.Add(vm.GetIface(virt.Id, i))
	}

	for i := 0; i < 60; i++ {
		ifaces, e := net.Interfaces()
		if e != nil {
			err = &errortypes.ReadError{
				errors.Wrap(e, "qemu: Failed to get network interfaces"),
			}
			return
		}

		for _, iface := range ifaces {
			if ifaceNames.Contains(iface.Name) {
				ifaceNames.Remove(iface.Name)
			}
		}

		if ifaceNames.Len() == 0 {
			break
		}

		time.Sleep(250 * time.Millisecond)
	}

	if ifaceNames.Len() != 0 {
		err = &errortypes.ReadError{
			errors.New("qemu: Failed to find network interfaces"),
		}
		return
	}

	if len(virt.NetworkAdapters) > 0 {
		iface := vm.GetIface(virt.Id, 0)
		adapter := virt.NetworkAdapters[0]

		err = utils.Exec("", "ip", "link", "set", iface, "up")
		if err != nil {
			PowerOff(db, virt)
			return
		}

		output, e := utils.ExecCombinedOutput(
			"", "brctl", "addif", adapter.HostInterface, iface)
		if e != nil {
			if !strings.Contains(output, "already a member of a bridge") {
				err = e
				PowerOff(db, virt)
				return
			}
		}
	}

	return
}

func writeService(virt *vm.VirtualMachine) (err error) {
	unitPath := paths.GetUnitPath(virt.Id)

	qm, err := NewQemu(virt)
	if err != nil {
		return
	}

	output, err := qm.Marshal()
	if err != nil {
		return
	}

	err = utils.CreateWrite(unitPath, output, 0644)
	if err != nil {
		return
	}

	err = systemd.Reload()
	if err != nil {
		return
	}

	return
}

func Create(db *database.Database, inst *instance.Instance,
	virt *vm.VirtualMachine) (err error) {

	vmPath := paths.GetVmPath(virt.Id)
	unitName := paths.GetUnitName(virt.Id)

	logrus.WithFields(logrus.Fields{
		"id": virt.Id.Hex(),
	}).Info("qemu: Creating virtual machine")

	virt.State = vm.Provisioning
	err = virt.Commit(db)
	if err != nil {
		return
	}

	err = utils.ExistsMkdir(settings.Hypervisor.LibPath, 0755)
	if err != nil {
		return
	}

	err = utils.ExistsMkdir(vmPath, 0755)
	if err != nil {
		return
	}

	dsk, err := disk.GetInstanceIndex(db, inst.Id, "0")
	if err != nil {
		if _, ok := err.(*database.NotFoundError); ok {
			dsk = nil
			err = nil
		} else {
			return
		}
	}

	if dsk == nil {
		dsk = &disk.Disk{
			Id:             bson.NewObjectId(),
			Name:           inst.Name,
			State:          disk.Available,
			Node:           node.Self.Id,
			Organization:   inst.Organization,
			Instance:       inst.Id,
			SourceInstance: inst.Id,
			Image:          virt.Image,
			Index:          "0",
			Size:           10,
		}

		err = data.WriteImage(db, virt.Image, dsk.Id)
		if err != nil {
			return
		}

		err = dsk.Insert(db)
		if err != nil {
			return
		}
	}

	virt.Disks = append(virt.Disks, &vm.Disk{
		Index: 0,
		Path:  paths.GetDiskPath(dsk.Id),
	})

	err = cloudinit.Write(db, virt.Id)
	if err != nil {
		return
	}

	err = writeService(virt)
	if err != nil {
		return
	}

	virt.State = vm.Starting
	err = virt.Commit(db)
	if err != nil {
		return
	}

	err = systemd.Start(unitName)
	if err != nil {
		return
	}

	err = Wait(db, virt)
	if err != nil {
		return
	}

	err = NetworkConf(db, virt)
	if err != nil {
		return
	}

	return
}

func Destroy(db *database.Database, virt *vm.VirtualMachine) (err error) {
	vmPath := paths.GetVmPath(virt.Id)
	unitName := paths.GetUnitName(virt.Id)
	unitPath := paths.GetUnitPath(virt.Id)
	sockPath := paths.GetSockPath(virt.Id)
	guestPath := paths.GetGuestPath(virt.Id)
	pidPath := paths.GetPidPath(virt.Id)

	logrus.WithFields(logrus.Fields{
		"id": virt.Id.Hex(),
	}).Info("qemu: Destroying virtual machine")

	exists, err := utils.Exists(unitPath)
	if err != nil {
		return
	}

	if exists {
		err = systemd.Stop(unitName)
		if err != nil {
			return
		}
	}

	time.Sleep(3 * time.Second)

	for i, dsk := range virt.Disks {
		ds, e := disk.Get(db, dsk.GetId())
		if e != nil {
			err = e
			return
		}

		if i == 0 && ds.SourceInstance == virt.Id {
			err = disk.Delete(db, ds.Id)
			if err != nil {
				return
			}
		} else {
			err = disk.Detach(db, dsk.GetId())
			if err != nil {
				return
			}
		}
	}

	err = utils.RemoveAll(vmPath)
	if err != nil {
		return
	}

	err = utils.RemoveAll(unitPath)
	if err != nil {
		return
	}

	err = utils.RemoveAll(sockPath)
	if err != nil {
		return
	}

	err = utils.RemoveAll(guestPath)
	if err != nil {
		return
	}

	err = utils.RemoveAll(pidPath)
	if err != nil {
		return
	}

	err = utils.RemoveAll(paths.GetInitPath(virt.Id))
	if err != nil {
		return
	}

	err = utils.RemoveAll(unitPath)
	if err != nil {
		return
	}

	return
}

func PowerOn(db *database.Database, virt *vm.VirtualMachine) (err error) {
	unitName := paths.GetUnitName(virt.Id)

	logrus.WithFields(logrus.Fields{
		"id": virt.Id.Hex(),
	}).Info("qemu: Stopping virtual machine")

	err = cloudinit.Write(db, virt.Id)
	if err != nil {
		return
	}

	err = writeService(virt)
	if err != nil {
		return
	}

	err = systemd.Start(unitName)
	if err != nil {
		return
	}

	err = Wait(db, virt)
	if err != nil {
		return
	}

	err = NetworkConf(db, virt)
	if err != nil {
		return
	}

	return
}

func PowerOff(db *database.Database, virt *vm.VirtualMachine) (err error) {
	unitName := paths.GetUnitName(virt.Id)

	logrus.WithFields(logrus.Fields{
		"id": virt.Id.Hex(),
	}).Info("qemu: Starting virtual machine")

	for i := 0; i < 10; i++ {
		err = qms.Shutdown(virt.Id)
		if err == nil {
			break
		}

		time.Sleep(500 * time.Millisecond)
	}

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"id":    virt.Id.Hex(),
			"error": err,
		}).Error("qemu: Power off virtual machine error")
		err = nil
	} else {
		for i := 0; i < settings.Hypervisor.StopTimeout; i++ {
			vrt, e := GetVmInfo(virt.Id, false)
			if e != nil {
				err = e
				return
			}

			if vrt == nil || vrt.State == vm.Stopped ||
				vrt.State == vm.Failed {

				if vrt != nil {
					err = vrt.Commit(db)
					if err != nil {
						return
					}
				}

				return
			}

			if i != 0 && i%3 == 0 {
				qms.Shutdown(virt.Id)
			}

			time.Sleep(1 * time.Second)
		}
	}

	logrus.WithFields(logrus.Fields{
		"id": virt.Id.Hex(),
	}).Warning("qemu: Force power off virtual machine")

	err = systemd.Stop(unitName)
	if err != nil {
		return
	}

	time.Sleep(3 * time.Second)

	return
}

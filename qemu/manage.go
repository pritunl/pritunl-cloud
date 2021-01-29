package qemu

import (
	"encoding/json"
	"io/ioutil"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/cloudinit"
	"github.com/pritunl/pritunl-cloud/constants"
	"github.com/pritunl/pritunl-cloud/data"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/iproute"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/qmp"
	"github.com/pritunl/pritunl-cloud/qms"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/store"
	"github.com/pritunl/pritunl-cloud/systemd"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/sirupsen/logrus"
)

var (
	serviceReg = regexp.MustCompile("pritunl_cloud_([a-z0-9]+).service")
)

type InfoCache struct {
	Timestamp time.Time
	Virt      *vm.VirtualMachine
}

func GetVmInfo(vmId primitive.ObjectID, queryQms, force bool) (
	virt *vm.VirtualMachine, err error) {

	refreshRate := time.Duration(
		settings.Hypervisor.RefreshRate) * time.Second

	virtStore, ok := store.GetVirt(vmId)
	if !ok {
		unitPath := paths.GetUnitPath(vmId)

		unitData, e := ioutil.ReadFile(unitPath)
		if e != nil {
			err = &errortypes.ReadError{
				errors.Wrap(e, "qemu: Failed to read service"),
			}
			_ = ForcePowerOffErr(virt, err)
			return
		}

		virt = &vm.VirtualMachine{}
		for _, line := range strings.Split(string(unitData), "\n") {
			if !strings.HasPrefix(line, "PritunlData=") &&
				!strings.HasPrefix(line, "# PritunlData=") {

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
				_ = ForcePowerOffErr(virt, err)
				return
			}

			break
		}

		if virt.Id.IsZero() {
			virt = nil
			return
		}

		_ = UpdateVmState(virt)
	} else {
		virt = &virtStore.Virt

		if force || virt.State != vm.Running ||
			time.Since(virtStore.Timestamp) > 6*time.Second {

			_ = UpdateVmState(virt)
		}
	}

	if virt.State == vm.Running && queryQms {
		virt.DisksAvailable = true
		disksUpdated := false
		disksStore, ok := store.GetDisks(vmId)

		if !ok || time.Since(disksStore.Timestamp) > refreshRate {
			for i := 0; i < 10; i++ {
				if virt.State == vm.Running {
					disks, e := qms.GetDisks(vmId)
					if e != nil {
						if i < 9 {
							time.Sleep(300 * time.Millisecond)
							_ = UpdateVmState(virt)
							continue
						}

						logrus.WithFields(logrus.Fields{
							"instance_id": vmId.Hex(),
							"error":       e,
						}).Error("qemu: Failed to get VM disk state")

						virt.DisksAvailable = false

						e = nil

						break
					}

					virt.Disks = disks
					store.SetDisks(vmId, disks)
					disksUpdated = true
				}

				break
			}
		}

		if ok && !disksUpdated {
			disks := []*vm.Disk{}
			for _, dsk := range disksStore.Disks {
				disks = append(disks, dsk.Copy())
			}
			virt.Disks = disks
		}
	}

	if virt.State == vm.Running && queryQms && node.Self.UsbPassthrough {
		virt.UsbDevicesAvailable = true
		usbsUpdated := false
		usbsStore, ok := store.GetUsbs(vmId)
		if !ok || time.Since(usbsStore.Timestamp) > refreshRate {
			for i := 0; i < 10; i++ {
				if virt.State == vm.Running {
					usbs, e := qms.GetUsbDevices(vmId)
					if e != nil {
						if i < 9 {
							time.Sleep(300 * time.Millisecond)
							_ = UpdateVmState(virt)
							continue
						}

						logrus.WithFields(logrus.Fields{
							"instance_id": vmId.Hex(),
							"error":       e,
						}).Error("qemu: Failed to get VM usb state")

						virt.UsbDevicesAvailable = false

						e = nil

						break
					}

					virt.UsbDevices = usbs
					store.SetUsbs(vmId, usbs)
					usbsUpdated = true
				}

				break
			}
		}

		if ok && !usbsUpdated {
			usbs := []*vm.UsbDevice{}
			for _, usb := range usbsStore.Usbs {
				usbs = append(usbs, usb.Copy())
			}
			virt.UsbDevices = usbs
		}
	}

	addrStore, ok := store.GetAddress(virt.Id)
	if !ok {
		addr := ""
		addr6 := ""

		namespace := vm.GetNamespace(virt.Id, 0)
		ifaceExternal := vm.GetIfaceExternal(virt.Id, 0)
		ifaceExternal6 := vm.GetIfaceExternal(virt.Id, 1)

		nodeNetworkMode := node.Self.NetworkMode
		if nodeNetworkMode == "" {
			nodeNetworkMode = node.Dhcp
		}
		nodeNetworkMode6 := node.Self.NetworkMode6

		externalNetwork := true
		if nodeNetworkMode == node.Internal {
			externalNetwork = false
		}

		externalNetwork6 := false
		if nodeNetworkMode6 != "" && (nodeNetworkMode != nodeNetworkMode6 ||
			(nodeNetworkMode6 == node.Static)) {

			externalNetwork6 = true
		}

		if externalNetwork {
			address, address6, e := iproute.AddressGetIface(
				namespace, ifaceExternal)
			if e != nil {
				err = e
				_ = ForcePowerOffErr(virt, err)
				return
			}

			if address != nil {
				addr = address.Local
			}

			if address6 != nil {
				addr6 = address6.Local
			}
		}

		if externalNetwork6 {
			_, address6, e := iproute.AddressGetIface(
				namespace, ifaceExternal6)
			if e != nil {
				err = e
				_ = ForcePowerOffErr(virt, err)
				return
			}

			if address6 != nil {
				addr6 = address6.Local
			}
		}

		if len(virt.NetworkAdapters) > 0 {
			virt.NetworkAdapters[0].IpAddress = addr
			virt.NetworkAdapters[0].IpAddress6 = addr6
		}
		store.SetAddress(virt.Id, addr, addr6)
	} else {
		if len(virt.NetworkAdapters) > 0 {
			virt.NetworkAdapters[0].IpAddress = addrStore.Addr
			virt.NetworkAdapters[0].IpAddress6 = addrStore.Addr6
		}
	}

	return
}

func UpdateVmState(virt *vm.VirtualMachine) (err error) {
	unitName := paths.GetUnitName(virt.Id)
	state, timestamp, err := systemd.GetState(unitName)
	if err != nil {
		return
	}

	switch state {
	case "active":
		virt.State = vm.Running
		break
	case "deactivating":
		virt.State = vm.Running
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

	virt.Timestamp = timestamp

	store.SetVirt(virt.Id, virt)

	return
}

func SetVmState(virt *vm.VirtualMachine, state string) {
	virt.State = state
	store.SetVirt(virt.Id, virt)
}

func GetVms(db *database.Database,
	instMap map[primitive.ObjectID]*instance.Instance) (
	virts []*vm.VirtualMachine, err error) {

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
		if match == nil || len(match) != 2 {
			continue
		}

		vmId, err := primitive.ObjectIDFromHex(match[1])
		if err != nil {
			continue
		}

		waiter.Add(1)
		go func() {
			defer waiter.Done()

			virt, e := GetVmInfo(vmId, true, false)
			if e != nil {
				err = e
				return
			}

			if virt != nil {
				inst := instMap[vmId]
				if inst != nil && inst.VmState == vm.Running &&
					(virt.State == vm.Stopped || virt.State == vm.Failed) {

					inst.State = instance.Cleanup
					e = virt.CommitState(db, instance.Cleanup)
				} else {
					e = virt.Commit(db)
				}
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

	waiter.Wait()

	return
}

func Wait(db *database.Database, virt *vm.VirtualMachine) (err error) {
	unitName := paths.GetUnitName(virt.Id)

	for i := 0; i < settings.Hypervisor.StartTimeout; i++ {
		err = UpdateVmState(virt)
		if err != nil {
			return
		}

		if virt.State == vm.Running {
			break
		}

		time.Sleep(1 * time.Second)
	}

	if virt.State != vm.Running {
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

func Create(db *database.Database, inst *instance.Instance,
	virt *vm.VirtualMachine) (err error) {

	vmPath := paths.GetVmPath(virt.Id)
	unitName := paths.GetUnitName(virt.Id)

	if constants.Interrupt {
		return
	}

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

	err = utils.ExistsMkdir(settings.Hypervisor.RunPath, 0755)
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
			Id:               primitive.NewObjectID(),
			Name:             inst.Name,
			State:            disk.Available,
			Node:             node.Self.Id,
			Organization:     inst.Organization,
			Instance:         inst.Id,
			SourceInstance:   inst.Id,
			Image:            virt.Image,
			Backing:          inst.ImageBacking,
			Index:            "0",
			Size:             inst.InitDiskSize,
			DeleteProtection: inst.DeleteProtection,
		}

		backingImage, e := data.WriteImage(db, virt.Image, dsk.Id,
			inst.InitDiskSize, inst.ImageBacking)
		if e != nil {
			err = e
			return
		}

		dsk.BackingImage = backingImage

		err = dsk.Insert(db)
		if err != nil {
			return
		}

		_ = event.PublishDispatch(db, "disk.change")

		virt.Disks = append(virt.Disks, &vm.Disk{
			Id:    dsk.Id,
			Index: 0,
			Path:  paths.GetDiskPath(dsk.Id),
		})
	}

	err = cloudinit.Write(db, inst, virt, true)
	if err != nil {
		return
	}

	err = writeOvmfVars(virt)
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

	if virt.Vnc {
		err = qmp.VncPassword(virt.Id, inst.VncPassword)
		if err != nil {
			return
		}
	}

	err = NetworkConf(db, virt)
	if err != nil {
		return
	}

	store.RemVirt(virt.Id)
	store.RemDisks(virt.Id)

	return
}

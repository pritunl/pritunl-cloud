package deploy

import (
	"sort"
	"strings"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/arp"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/netconf"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/permission"
	"github.com/pritunl/pritunl-cloud/qemu"
	"github.com/pritunl/pritunl-cloud/qmp"
	"github.com/pritunl/pritunl-cloud/qms"
	"github.com/pritunl/pritunl-cloud/state"
	"github.com/pritunl/pritunl-cloud/store"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/pritunl/pritunl-cloud/vpc"
	"github.com/sirupsen/logrus"
)

var (
	instancesLock = utils.NewMultiTimeoutLock(5 * time.Minute)
	limiter       = utils.NewLimiter(5)
)

type Instances struct {
	stat *state.State
}

func (s *Instances) create(inst *instance.Instance) {
	if !limiter.Acquire() {
		return
	}

	acquired, lockId := instancesLock.LockOpenTimeout(
		inst.Id.Hex(), 10*time.Minute)
	if !acquired {
		limiter.Release()
		return
	}

	go func() {
		defer func() {
			time.Sleep(3 * time.Second)
			instancesLock.Unlock(inst.Id.Hex(), lockId)
			limiter.Release()
		}()

		db := database.GetDatabase()
		defer db.Close()

		err := qemu.Create(db, inst, inst.Virt)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"instance_id": inst.Id.Hex(),
				"error":       err,
			}).Error("deploy: Failed to create instance")

			err = instance.SetAction(db, inst.Id, instance.Stop)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"instance_id": inst.Id.Hex(),
					"error":       err,
				}).Error("deploy: Failed to set instance state")

				qemu.PowerOff(db, inst.Virt)

				return
			}

			return
		}

		event.PublishDispatch(db, "instance.change")
	}()
}

func (s *Instances) start(inst *instance.Instance) {
	if !limiter.Acquire() {
		return
	}

	acquired, lockId := instancesLock.LockOpen(inst.Id.Hex())
	if !acquired {
		limiter.Release()
		return
	}

	go func() {
		defer func() {
			time.Sleep(3 * time.Second)
			instancesLock.Unlock(inst.Id.Hex(), lockId)
			limiter.Release()
		}()

		db := database.GetDatabase()
		defer db.Close()

		if inst.Restart || inst.RestartBlockIp {
			inst.Restart = false
			inst.RestartBlockIp = false
			err := inst.CommitFields(db,
				set.NewSet("restart", "restart_block_ip"))
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"instance_id": inst.Id.Hex(),
					"error":       err,
				}).Error("deploy: Failed to commit instance")
				return
			}
		}

		err := qemu.PowerOn(db, inst, inst.Virt)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"instance_id": inst.Id.Hex(),
				"error":       err,
			}).Error("deploy: Failed to start instance")

			err = instance.SetAction(db, inst.Id, instance.Stop)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"instance_id": inst.Id.Hex(),
					"error":       err,
				}).Error("deploy: Failed to set instance state")

				qemu.PowerOff(db, inst.Virt)

				return
			}

			return
		}

		event.PublishDispatch(db, "instance.change")
	}()
}

func (s *Instances) cleanup(inst *instance.Instance) {
	if !limiter.Acquire() {
		return
	}

	acquired, lockId := instancesLock.LockOpen(inst.Id.Hex())
	if !acquired {
		limiter.Release()
		return
	}

	go func() {
		defer func() {
			time.Sleep(3 * time.Second)
			instancesLock.Unlock(inst.Id.Hex(), lockId)
			limiter.Release()
		}()

		db := database.GetDatabase()
		defer db.Close()

		qemu.Cleanup(db, inst.Virt)

		err := instance.SetAction(db, inst.Id, instance.Stop)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"instance_id": inst.Id.Hex(),
				"error":       err,
			}).Error("deploy: Failed to update instance")
			return
		}

		event.PublishDispatch(db, "instance.change")
	}()
}

func (s *Instances) stop(inst *instance.Instance) {
	if !limiter.Acquire() {
		return
	}

	acquired, lockId := instancesLock.LockOpen(inst.Id.Hex())
	if !acquired {
		limiter.Release()
		return
	}

	go func() {
		defer func() {
			time.Sleep(3 * time.Second)
			instancesLock.Unlock(inst.Id.Hex(), lockId)
			limiter.Release()
		}()

		db := database.GetDatabase()
		defer db.Close()

		err := qemu.PowerOff(db, inst.Virt)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"instance_id": inst.Id.Hex(),
				"error":       err,
			}).Error("deploy: Failed to stop instance")
			return
		}

		event.PublishDispatch(db, "instance.change")
	}()
}

func (s *Instances) restart(inst *instance.Instance) {
	if !limiter.Acquire() {
		return
	}

	acquired, lockId := instancesLock.LockOpen(inst.Id.Hex())
	if !acquired {
		limiter.Release()
		return
	}

	go func() {
		defer func() {
			time.Sleep(3 * time.Second)
			instancesLock.Unlock(inst.Id.Hex(), lockId)
			limiter.Release()
		}()

		db := database.GetDatabase()
		defer db.Close()

		err := qemu.PowerOff(db, inst.Virt)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"instance_id": inst.Id.Hex(),
				"error":       err,
			}).Error("deploy: Failed to restart instance")
			return
		}

		time.Sleep(1 * time.Second)

		if inst.Restart || inst.RestartBlockIp {
			inst.Restart = false
			inst.RestartBlockIp = false
			err = inst.CommitFields(db,
				set.NewSet("restart", "restart_block_ip"))
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"instance_id": inst.Id.Hex(),
					"error":       err,
				}).Error("deploy: Failed to commit instance")
				return
			}
		}

		err = qemu.PowerOn(db, inst, inst.Virt)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"instance_id": inst.Id.Hex(),
				"error":       err,
			}).Error("deploy: Failed to restart instance")

			err = instance.SetAction(db, inst.Id, instance.Stop)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"instance_id": inst.Id.Hex(),
					"error":       err,
				}).Error("deploy: Failed to set instance state")

				qemu.PowerOff(db, inst.Virt)

				return
			}

			return
		}

		inst.Action = instance.Start
		err = inst.CommitFields(db, set.NewSet("action"))
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"instance_id": inst.Id.Hex(),
				"error":       err,
			}).Error("deploy: Failed to commit instance")
			return
		}

		event.PublishDispatch(db, "instance.change")
	}()
}

func (s *Instances) destroy(inst *instance.Instance) {
	if !limiter.Acquire() {
		return
	}

	acquired, lockId := instancesLock.LockOpen(inst.Id.Hex())
	if !acquired {
		limiter.Release()
		return
	}

	go func() {
		defer func() {
			time.Sleep(3 * time.Second)
			instancesLock.Unlock(inst.Id.Hex(), lockId)
			limiter.Release()
		}()

		db := database.GetDatabase()
		defer db.Close()

		_, err := instance.Get(db, inst.Id)
		if err != nil {
			if _, ok := err.(*database.NotFoundError); ok {
				err = nil
			}
			return
		}

		err = qemu.Destroy(db, inst.Virt)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"instance_id": inst.Id.Hex(),
				"error":       err,
			}).Error("deploy: Failed to power off instance")
			return
		}

		err = netconf.Destroy(db, inst.Virt)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"instance_id": inst.Id.Hex(),
				"error":       err,
			}).Error("deploy: Failed to destroy netconf")
			return
		}

		err = instance.Remove(db, inst.Id)
		if err != nil {
			if _, ok := err.(*database.NotFoundError); ok {
				err = nil
			} else {
				logrus.WithFields(logrus.Fields{
					"instance_id": inst.Id.Hex(),
					"error":       err,
				}).Error("deploy: Failed to remove instance")
				return
			}
		}

		event.PublishDispatch(db, "instance.change")
		event.PublishDispatch(db, "disk.change")
	}()
}

func (s *Instances) diskAdd(inst *instance.Instance,
	virt *vm.VirtualMachine, addDisks vm.SortDisks) {

	acquired, lockId := instancesLock.LockOpen(inst.Id.Hex())
	if !acquired {
		return
	}

	go func() {
		defer func() {
			time.Sleep(3 * time.Second)
			instancesLock.Unlock(inst.Id.Hex(), lockId)
		}()

		db := database.GetDatabase()
		defer db.Close()

		sort.Sort(addDisks)

		for _, dsk := range addDisks {
			err := permission.InitDisk(virt, dsk)
			if err != nil {
				return
			}

			err = qmp.AddDisk(inst.Id, dsk)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"instance_id": inst.Id.Hex(),
					"disk_id":     dsk.Id.Hex(),
					"error":       err,
				}).Error("sync: Failed to add disk")
				return
			}
		}

		time.Sleep(200 * time.Millisecond)

		err := qemu.UpdateVmDisk(virt)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"instance_id": inst.Id.Hex(),
				"error":       err,
			}).Error("sync: Failed to update vm disk state")
		}

		event.PublishDispatch(db, "instance.change")
		event.PublishDispatch(db, "disk.change")
	}()
}

func (s *Instances) diskRemove(inst *instance.Instance,
	virt *vm.VirtualMachine, remDisks vm.SortDisks) {

	acquired, lockId := instancesLock.LockOpen(inst.Id.Hex())
	if !acquired {
		return
	}

	go func() {
		defer func() {
			time.Sleep(3 * time.Second)
			instancesLock.Unlock(inst.Id.Hex(), lockId)
		}()

		db := database.GetDatabase()
		defer db.Close()

		sort.Sort(sort.Reverse(remDisks))

		for _, dsk := range remDisks {
			e := qmp.RemoveDisk(inst.Id, dsk)
			if e != nil {
				logrus.WithFields(logrus.Fields{
					"instance_id": inst.Id.Hex(),
					"disk_id":     dsk.Id.Hex(),
					"error":       e,
				}).Error("sync: Failed to remove disk")
				return
			}
		}

		time.Sleep(200 * time.Millisecond)

		err := qemu.UpdateVmDisk(virt)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"instance_id": inst.Id.Hex(),
				"error":       err,
			}).Error("sync: Failed to update vm disk state")
		}

		event.PublishDispatch(db, "instance.change")
		event.PublishDispatch(db, "disk.change")
	}()
}

func (s *Instances) usbAdd(inst *instance.Instance, virt *vm.VirtualMachine,
	addUsbs []*vm.UsbDevice) {

	acquired, lockId := instancesLock.LockOpen(inst.Id.Hex())
	if !acquired {
		return
	}

	go func() {
		defer func() {
			time.Sleep(3 * time.Second)
			instancesLock.Unlock(inst.Id.Hex(), lockId)
		}()

		db := database.GetDatabase()
		defer db.Close()

		for _, device := range addUsbs {
			e := qms.AddUsb(inst.Id, device)
			if e != nil {
				logrus.WithFields(logrus.Fields{
					"instance_id": inst.Id.Hex(),
					"usb_address": device.Address,
					"usb_bus":     device.Bus,
					"usb_product": device.Product,
					"usb_vendor":  device.Vendor,
					"error":       e,
				}).Error("sync: Failed to add usb")
				return
			}
		}

		time.Sleep(200 * time.Millisecond)

		err := qemu.UpdateVmUsb(virt)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"instance_id": inst.Id.Hex(),
				"error":       err,
			}).Error("sync: Failed to update vm usb state")
		}

		event.PublishDispatch(db, "instance.change")
	}()
}

func (s *Instances) usbRemove(inst *instance.Instance,
	virt *vm.VirtualMachine, remUsbs []*vm.UsbDevice) {

	acquired, lockId := instancesLock.LockOpen(inst.Id.Hex())
	if !acquired {
		return
	}

	go func() {
		defer func() {
			time.Sleep(3 * time.Second)
			instancesLock.Unlock(inst.Id.Hex(), lockId)
		}()

		db := database.GetDatabase()
		defer db.Close()

		for _, device := range remUsbs {
			err := qms.RemoveUsb(inst.Id, device)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"instance_id": inst.Id.Hex(),
					"usb_address": device.Address,
					"usb_bus":     device.Bus,
					"usb_product": device.Product,
					"usb_vendor":  device.Vendor,
					"error":       err,
				}).Error("sync: Failed to remove usb")
				return
			}
		}

		time.Sleep(200 * time.Millisecond)

		err := qemu.UpdateVmUsb(virt)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"instance_id": inst.Id.Hex(),
				"error":       err,
			}).Error("sync: Failed to update vm usb state")
		}

		event.PublishDispatch(db, "instance.change")
	}()
}

func (s *Instances) diff(db *database.Database,
	inst *instance.Instance) (err error) {

	curVirt := s.stat.GetVirt(inst.Id)
	if curVirt == nil {
		err = &errortypes.ReadError{
			errors.New("deploy: Failed to load virt"),
		}
		logrus.WithFields(logrus.Fields{
			"instance_id": inst.Id.Hex(),
			"error":       err,
		}).Error("deploy: Failed to load virt")
		err = nil
		return
	}

	changed := inst.Changed(curVirt)
	addDisks, remDisks := inst.DiskChanged(curVirt)
	addUsbs, remUsbs := inst.UsbChanged(curVirt)

	if instancesLock.Locked(inst.Id.Hex()) {
		return
	}

	if changed && !inst.Restart {
		inst.Restart = true
		err = inst.CommitFields(db, set.NewSet("restart"))
		if err != nil {
			return
		}
	} else if !changed && inst.Restart {
		inst.Restart = false
		err = inst.CommitFields(db, set.NewSet("restart"))
		if err != nil {
			return
		}
	}

	if len(remDisks) > 0 {
		s.diskRemove(inst, curVirt, remDisks)
	}

	if len(addDisks) > 0 {
		s.diskAdd(inst, curVirt, addDisks)
	}

	if len(remUsbs) > 0 {
		s.usbRemove(inst, curVirt, remUsbs)
	}

	if len(addUsbs) > 0 {
		s.usbAdd(inst, curVirt, addUsbs)
	}

	return
}

func (s *Instances) check(inst *instance.Instance, namespaces set.Set) (
	valid bool, err error) {

	namespace := vm.GetNamespace(inst.Id, 0)
	if !namespaces.Contains(namespace) {
		logrus.WithFields(logrus.Fields{
			"instance_id":   inst.Id.Hex(),
			"net_namespace": namespace,
		}).Error("deploy: Instance missing namespace")
		return
	}

	valid = true

	return
}

func (s *Instances) routes(inst *instance.Instance) (err error) {
	acquired, lockId := instancesLock.LockOpen(inst.Id.Hex())
	if !acquired {
		return
	}

	go func() {
		defer func() {
			instancesLock.Unlock(inst.Id.Hex(), lockId)
		}()

		namespace := vm.GetNamespace(inst.Id, 0)

		vc := s.stat.Vpc(inst.Vpc)
		if vc == nil {
			err = &errortypes.NotFoundError{
				errors.New("deploy: Instance vpc not found"),
			}
			logrus.WithFields(logrus.Fields{
				"instance_id":   inst.Id.Hex(),
				"net_namespace": namespace,
				"error":         err,
			}).Error("deploy: Failed to deploy instance routes")
			return
		}

		curRoutes := set.NewSet()
		curRoutes6 := set.NewSet()
		newRoutes := set.NewSet()
		newRoutes6 := set.NewSet()

		var routes []vpc.Route
		var routes6 []vpc.Route

		routesStore, ok := store.GetRoutes(inst.Id)
		if !ok {
			routes, routes6, err = qemu.GetRoutes(inst.Id)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"instance_id":   inst.Id.Hex(),
					"net_namespace": namespace,
					"error":         err,
				}).Error("deploy: Failed to deploy instance routes")
				return
			}

			if routes == nil || routes6 == nil {
				return
			}

			store.SetRoutes(inst.Id, routes, routes6)
		} else {
			routes = routesStore.Routes
			routes6 = routesStore.Routes6
		}

		for _, route := range routes {
			curRoutes.Add(route)
		}

		for _, route := range routes6 {
			curRoutes6.Add(route)
		}

		if vc.Routes != nil {
			for _, route := range vc.Routes {
				if !strings.Contains(route.Destination, ":") {
					newRoutes.Add(*route)
				} else {
					newRoutes6.Add(*route)
				}
			}
		}

		changed := false
		addRoutes := newRoutes.Copy()
		addRoutes6 := newRoutes6.Copy()
		remRoutes := curRoutes.Copy()
		remRoutes6 := curRoutes6.Copy()

		addRoutes.Subtract(curRoutes)
		addRoutes6.Subtract(curRoutes6)
		remRoutes.Subtract(newRoutes)
		remRoutes6.Subtract(newRoutes6)

		for routeInf := range remRoutes.Iter() {
			route := routeInf.(vpc.Route)
			changed = true

			utils.ExecCombinedOutputLogged(
				nil,
				"ip", "netns", "exec", namespace,
				"ip", "route",
				"del", route.Destination,
				"via", route.Target,
				"metric", "97",
			)
		}

		for routeInf := range remRoutes6.Iter() {
			route := routeInf.(vpc.Route)
			changed = true

			utils.ExecCombinedOutputLogged(
				nil,
				"ip", "netns", "exec", namespace,
				"ip", "-6", "route",
				"del", route.Destination,
				"via", route.Target,
				"metric", "97",
			)
		}

		for routeInf := range addRoutes.Iter() {
			route := routeInf.(vpc.Route)
			changed = true

			utils.ExecCombinedOutputLogged(
				[]string{
					"File exists",
				},
				"ip", "netns", "exec", namespace,
				"ip", "route",
				"add", route.Destination,
				"via", route.Target,
				"metric", "97",
			)
		}

		for routeInf := range addRoutes6.Iter() {
			route := routeInf.(vpc.Route)
			changed = true

			utils.ExecCombinedOutputLogged(
				[]string{
					"File exists",
				},
				"ip", "netns", "exec", namespace,
				"ip", "-6", "route",
				"add", route.Destination,
				"via", route.Target,
				"metric", "97",
			)
		}

		if changed {
			store.RemRoutes(inst.Id)
		}

		var curRecords set.Set

		recordsStore, ok := store.GetArp(inst.Id)
		if !ok {
			curRecords, err = arp.GetRecords(namespace)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"instance_id": inst.Id.Hex(),
					"error":       err,
				}).Error("deploy: Failed to deploy instance arp table")
				return
			}

			if routes == nil || routes6 == nil {
				return
			}

			store.SetArp(inst.Id, curRecords)
		} else {
			curRecords = recordsStore.Records
		}

		newRecords := s.stat.ArpRecords(namespace)

		if curRecords == nil || newRecords == nil {
			logrus.WithFields(logrus.Fields{
				"cur_records_nil": curRecords == nil,
				"new_records_nil": newRecords == nil,
			}).Error("deploy: Missing arp records")
			return
		}

		changed, err = arp.ApplyState(namespace, curRecords, newRecords)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"instance_id": inst.Id.Hex(),
				"error":       err,
			}).Error("deploy: Failed to deploy instance arp table")
			return
		}

		if changed {
			store.RemArp(inst.Id)
		}
	}()

	return
}

func (s *Instances) Deploy(db *database.Database) (err error) {
	instances := s.stat.Instances()
	namespaces := s.stat.Namespaces()

	namespacesSet := set.NewSet()
	for _, namespace := range namespaces {
		namespacesSet.Add(namespace)
	}

	cpuUnits := 0
	memoryUnits := 0.0

	for _, inst := range instances {
		curVirt := s.stat.GetVirt(inst.Id)

		if inst.Action == instance.Destroy {
			if inst.DeleteProtection {
				logrus.WithFields(logrus.Fields{
					"instance_id": inst.Id.Hex(),
				}).Info("deploy: Delete protection ignore instance destroy")

				if curVirt != nil && curVirt.State == vm.Running {
					inst.Action = instance.Start
				} else {
					inst.Action = instance.Stop
				}
				err = inst.CommitFields(db, set.NewSet("action"))
				if err != nil {
					return
				}

				event.PublishDispatch(db, "instance.change")
			} else {
				s.destroy(inst)
			}
			continue
		}

		cpuUnits += inst.Processors
		memoryUnits += float64(inst.Memory) / float64(1024)

		if curVirt == nil {
			if inst.Action == instance.Start {
				s.create(inst)
			}

			continue
		}

		switch inst.Action {
		case instance.Start:
			if curVirt.State == vm.Stopped || curVirt.State == vm.Failed {
				dsks := s.stat.GetInstaceDisks(inst.Id)

				for _, dsk := range dsks {
					if dsk.State != disk.Available {
						continue
					}
				}

				s.start(inst)
				continue
			}

			valid, e := s.check(inst, namespacesSet)
			if e != nil {
				err = e
				return
			}
			if !valid {
				continue
			}

			err = s.diff(db, inst)
			if err != nil {
				return
			}

			err = s.routes(inst)
			if err != nil {
				return
			}

			break
		case instance.Cleanup:
			s.cleanup(inst)
			continue
		case instance.Stop:
			if curVirt.State == vm.Running {
				s.stop(inst)
				continue
			}
			break
		case instance.Restart:
			if curVirt.State == vm.Running {
				dsks := s.stat.GetInstaceDisks(inst.Id)

				for _, dsk := range dsks {
					if dsk.State != disk.Available {
						continue
					}
				}

				s.restart(inst)
				continue
			} else if curVirt.State == vm.Stopped ||
				curVirt.State == vm.Failed {

				inst.Action = instance.Start
				err = inst.CommitFields(db, set.NewSet("action"))
				if err != nil {
					return
				}

				continue
			}
			break
		}
	}

	node.Self.CpuUnitsRes = cpuUnits
	node.Self.MemoryUnitsRes = memoryUnits

	return
}

func NewInstances(stat *state.State) *Instances {
	return &Instances{
		stat: stat,
	}
}

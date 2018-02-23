package sync

import (
	"github.com/Sirupsen/logrus"
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/pritunl-cloud/bridge"
	"github.com/pritunl/pritunl-cloud/constants"
	"github.com/pritunl/pritunl-cloud/data"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/iptables"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/qemu"
	"github.com/pritunl/pritunl-cloud/qms"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"gopkg.in/mgo.v2/bson"
	"time"
)

var (
	busy = utils.NewMultiLock()
)

func vmUpdate() (err error) {
	db := database.GetDatabase()
	defer db.Close()

	disks, err := disk.GetNode(db, node.Self.Id)
	if err != nil {
		return
	}

	curVirts, err := qemu.GetVms(db)
	if err != nil {
		return
	}

	curIds := set.NewSet()
	curVirtsMap := map[bson.ObjectId]*vm.VirtualMachine{}
	for _, virt := range curVirts {
		curIds.Add(virt.Id)
		curVirtsMap[virt.Id] = virt
	}

	availableDisks := []*disk.Disk{}
	for _, dsk := range disks {
		switch dsk.State {
		case disk.Provision:
			err = data.CreateDisk(db, dsk)
			if err != nil {
				return
			}

			dsk.State = disk.Available
			err = dsk.CommitFields(db, set.NewSet("state"))
			if err != nil {
				return
			}

			break
		case disk.Available:
			availableDisks = append(availableDisks, dsk)
			break
		case disk.Destroy:
			var curVirt *vm.VirtualMachine
			if dsk.Instance != "" {
				curVirt = curVirtsMap[dsk.Instance]
			}

			inUse := false
			if curVirt != nil {
				for _, vmDsk := range curVirt.Disks {
					if vmDsk.GetId() == dsk.Id {
						inUse = true
						break
					}
				}
			}

			if !inUse && !busy.Locked(dsk.Id.Hex()) {
				busy.Lock(dsk.Id.Hex())
				go func(dsk *disk.Disk) {
					defer func() {
						busy.Unlock(dsk.Id.Hex())
					}()

					db := database.GetDatabase()
					defer db.Close()

					e := dsk.Destroy(db)
					if e != nil {
						logrus.WithFields(logrus.Fields{
							"error": e,
						}).Error("sync: Failed to destroy disk")
						return
					}

					event.PublishDispatch(db, "disk.change")
				}(dsk)
			}

			break

		}
	}

	instances, err := instance.GetAllVirt(db, &bson.M{
		"node": node.Self.Id,
	}, availableDisks)

	err = iptables.UpdateState(db, instances)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("sync: Failed to update iptables, resetting state")
		for {
			err = iptables.Recover()
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Error("sync: Failed to recover iptables, retrying")
				continue
			}
			break
		}
		err = nil
		return
	}

	newIds := set.NewSet()
	for _, inst := range instances {
		newIds.Add(inst.Id)
		if !curIds.Contains(inst.Id) && !busy.Locked(inst.Id.Hex()) {
			busy.Lock(inst.Id.Hex())
			go func(inst *instance.Instance) {
				defer func() {
					time.Sleep(5 * time.Second)
					busy.Unlock(inst.Id.Hex())
				}()
				db := database.GetDatabase()
				defer db.Close()

				e := qemu.Create(db, inst, inst.Virt)
				if e != nil {
					logrus.WithFields(logrus.Fields{
						"error": e,
					}).Error("sync: Failed to create instance")
					return
				}
			}(inst)
		}
	}

	curIds.Subtract(newIds)
	for idInf := range curIds.Iter() {
		logrus.WithFields(logrus.Fields{
			"id": idInf.(bson.ObjectId),
		}).Info("sync: Unknown instance")
	}

	for _, inst := range instances {
		curVirt := curVirtsMap[inst.Id]
		if curVirt == nil {
			continue
		}

		switch inst.State {
		case instance.Start:
			if (curVirt.State == vm.Stopped || curVirt.State == vm.Failed) &&
				!busy.Locked(inst.Id.Hex()) {

				busy.Lock(inst.Id.Hex())
				go func(inst *instance.Instance) {
					defer func() {
						time.Sleep(3 * time.Second)
						busy.Unlock(inst.Id.Hex())
					}()
					db := database.GetDatabase()
					defer db.Close()

					e := qemu.PowerOn(db, inst.Virt)
					if e != nil {
						logrus.WithFields(logrus.Fields{
							"error": e,
						}).Error("sync: Failed to power on instance")
						return
					}
				}(inst)
				continue
			} else if inst.Changed(curVirt) && !inst.Restart {
				inst.Restart = true
				err = inst.CommitFields(db, set.NewSet("restart"))
				if err != nil {
					return
				}
			} else if !inst.Changed(curVirt) && inst.Restart {
				inst.Restart = false
				err = inst.CommitFields(db, set.NewSet("restart"))
				if err != nil {
					return
				}
			}

			addDisks, remDisks := inst.DiskChanged(curVirt)
			if len(addDisks) > 0 && !inst.Restart {
				inst.Restart = true
				err = inst.CommitFields(db, set.NewSet("restart"))
				if err != nil {
					return
				}
			}

			if len(remDisks) > 0 && !busy.Locked(inst.Id.Hex()) {
				busy.Lock(inst.Id.Hex())
				go func(inst *instance.Instance) {
					defer func() {
						time.Sleep(3 * time.Second)
						busy.Unlock(inst.Id.Hex())
					}()

					for _, dsk := range remDisks {
						e := qms.RemoveDisk(inst.Id, dsk)
						if e != nil {
							logrus.WithFields(logrus.Fields{
								"error": e,
							}).Error("sync: Failed to remove disk")
							return
						}
					}
				}(inst)
				continue
			}
			break
		case instance.Stop:
			if curVirt.State == vm.Running && !busy.Locked(inst.Id.Hex()) {
				busy.Lock(inst.Id.Hex())
				go func(inst *instance.Instance) {
					defer busy.Unlock(inst.Id.Hex())
					db := database.GetDatabase()
					defer db.Close()

					e := qemu.PowerOff(db, inst.Virt)
					if e != nil {
						logrus.WithFields(logrus.Fields{
							"error": e,
						}).Error("sync: Failed to power off instance")
						return
					}
				}(inst)
				continue
			}
			break
		case instance.Restart:
			if !busy.Locked(inst.Id.Hex()) {
				busy.Lock(inst.Id.Hex())
				go func(inst *instance.Instance) {
					defer busy.Unlock(inst.Id.Hex())

					db := database.GetDatabase()
					defer db.Close()

					e := qemu.PowerOff(db, inst.Virt)
					if e != nil {
						logrus.WithFields(logrus.Fields{
							"error": e,
						}).Error("sync: Failed to power off instance")
						return
					}

					time.Sleep(1 * time.Second)

					e = qemu.PowerOn(db, inst.Virt)
					if e != nil {
						logrus.WithFields(logrus.Fields{
							"error": e,
						}).Error("sync: Failed to power on instance")
						return
					}

					inst.State = instance.Start
					err = inst.CommitFields(db, set.NewSet("state"))
					if err != nil {
						return
					}
				}(inst)
				continue
			}
			break
		case instance.Destroy:
			if !busy.Locked(inst.Id.Hex()) {
				busy.Lock(inst.Id.Hex())
				go func(inst *instance.Instance) {
					defer busy.Unlock(inst.Id.Hex())

					db := database.GetDatabase()
					defer db.Close()

					e := qemu.Destroy(db, inst.Virt)
					if e != nil {
						logrus.WithFields(logrus.Fields{
							"error": e,
						}).Error("sync: Failed to power off instance")
						return
					}

					e = instance.Remove(db, inst.Id)
					if e != nil {
						logrus.WithFields(logrus.Fields{
							"error": e,
						}).Error("sync: Failed to remove instance")
						return
					}

					event.PublishDispatch(db, "disk.change")
				}(inst)
				continue
			}
			break
		case instance.Snapshot:
			if !busy.Locked(inst.Id.Hex()) {
				busy.Lock(inst.Id.Hex())
				go func(inst *instance.Instance) {
					defer busy.Unlock(inst.Id.Hex())

					db := database.GetDatabase()
					defer db.Close()

					e := data.CreateSnapshot(db, inst.Virt)
					if e != nil {
						logrus.WithFields(logrus.Fields{
							"error": e,
						}).Error("sync: Failed to snapshot instance")
						return
					}

					if curVirt.State == vm.Running {
						inst.State = instance.Start
					} else {
						inst.State = instance.Stop
					}
					e = inst.CommitFields(db, set.NewSet("state"))
					if e != nil {
						logrus.WithFields(logrus.Fields{
							"error": e,
						}).Error("sync: Failed to update instance")
						return
					}
				}(inst)
				continue
			}
			break
		}
	}

	return
}

func syncNodeFirewall() {
	db := database.GetDatabase()
	defer db.Close()

	err := iptables.UpdateState(db, []*instance.Instance{})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("sync: Failed to update iptables, resetting state")
		for {
			err = iptables.Recover()
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Error("sync: Failed to recover iptables, retrying")
				continue
			}
			break
		}
	}
}

func vmRunner() {
	time.Sleep(1 * time.Second)

	for {
		time.Sleep(1 * time.Second)
		if !node.Self.IsHypervisor() {
			syncNodeFirewall()
			continue
		}

		err := bridge.Configure()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("sync: Failed to configure bridge")

			time.Sleep(1 * time.Second)

			continue
		}

		break
	}

	logrus.WithFields(logrus.Fields{
		"production": constants.Production,
		"bridge":     bridge.BridgeName,
	}).Info("bridge: Starting hypervisor")

	for {
		time.Sleep(1 * time.Second)
		if !node.Self.IsHypervisor() {
			syncNodeFirewall()
			continue
		}

		err := vmUpdate()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("sync: Failed to update vm")
			continue
		}
	}
}

func initVm() {
	go vmRunner()
}

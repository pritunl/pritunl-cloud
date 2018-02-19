package sync

import (
	"github.com/Sirupsen/logrus"
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/qemu"
	"github.com/pritunl/pritunl-cloud/vm"
	"gopkg.in/mgo.v2/bson"
	"sync"
	"time"
)

var (
	busy     = set.NewSet()
	busyLock = sync.Mutex{}
)

func vmUpdate() (err error) {
	db := database.GetDatabase()
	defer db.Close()

	virts, err := qemu.GetVms(db)
	if err != nil {
		return
	}

	curIds := set.NewSet()
	virtsMap := map[bson.ObjectId]*vm.VirtualMachine{}
	for _, virt := range virts {
		curIds.Add(virt.Id)
		virtsMap[virt.Id] = virt
	}

	instances, err := instance.GetAll(db, &bson.M{
		"node": node.Self.Id,
	})

	newIds := set.NewSet()
	for _, inst := range instances {
		newIds.Add(inst.Id)
		if !curIds.Contains(inst.Id) && !busy.Contains(inst.Id) {
			busyLock.Lock()
			busy.Add(inst.Id)
			busyLock.Unlock()
			go func(inst *instance.Instance) {
				defer func() {
					busyLock.Lock()
					busy.Remove(inst.Id)
					busyLock.Unlock()
				}()
				db := database.GetDatabase()
				defer db.Close()

				e := qemu.Create(db, inst.GetVm())
				time.Sleep(5 * time.Second)
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
		virt := virtsMap[inst.Id]
		if virt == nil {
			continue
		}

		switch inst.State {
		case instance.Running:
			if virt.State == vm.Stopped && !busy.Contains(inst.Id) {
				busyLock.Lock()
				busy.Add(inst.Id)
				busyLock.Unlock()
				go func(inst *instance.Instance) {
					defer func() {
						busyLock.Lock()
						busy.Remove(inst.Id)
						busyLock.Unlock()
					}()
					db := database.GetDatabase()
					defer db.Close()

					e := qemu.PowerOn(db, inst.GetVm())
					if e != nil {
						logrus.WithFields(logrus.Fields{
							"error": e,
						}).Error("sync: Failed to power on instance")
						return
					}

					time.Sleep(3 * time.Second)
				}(inst)
				continue
			}
			break
		case instance.Stopped:
			if virt.State == vm.Running && !busy.Contains(inst.Id) {
				busyLock.Lock()
				busy.Add(inst.Id)
				busyLock.Unlock()
				go func(inst *instance.Instance) {
					defer func() {
						busyLock.Lock()
						busy.Remove(inst.Id)
						busyLock.Unlock()
					}()
					db := database.GetDatabase()
					defer db.Close()

					e := qemu.PowerOff(db, inst.GetVm())
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
		case instance.Updating:
			if !busy.Contains(inst.Id) {
				if virt.State != vm.Stopped {
					busyLock.Lock()
					busy.Add(inst.Id)
					busyLock.Unlock()
					go func(inst *instance.Instance) {
						defer func() {
							busyLock.Lock()
							busy.Remove(inst.Id)
							busyLock.Unlock()
						}()
						db := database.GetDatabase()
						defer db.Close()

						e := qemu.PowerOff(db, inst.GetVm())
						if e != nil {
							logrus.WithFields(logrus.Fields{
								"error": e,
							}).Error("sync: Failed to power off instance")
							return
						}
					}(inst)
					continue
				}

				if inst.Changed(virt) {
					busyLock.Lock()
					busy.Add(inst.Id)
					busyLock.Unlock()

					logrus.WithFields(logrus.Fields{
						"id":             virt.Id.Hex(),
						"memory_old":     virt.Memory,
						"memory":         inst.Memory,
						"processors_old": virt.Processors,
						"processors":     inst.Processors,
					}).Info("sync: Resizing virtual machine")

					go func(inst *instance.Instance) {
						defer func() {
							busyLock.Lock()
							busy.Remove(inst.Id)
							busyLock.Unlock()
						}()
						db := database.GetDatabase()
						defer db.Close()

						e := qemu.Update(db, inst.GetVm())
						if e != nil {
							logrus.WithFields(logrus.Fields{
								"error": e,
							}).Error("sync: Failed to update instance")
							return
						}

						time.Sleep(5 * time.Second)

						inst.State = instance.Stopped
						err = inst.CommitFields(db, set.NewSet("state"))
						if err != nil {
							return
						}
					}(inst)
				} else {
					inst.State = instance.Stopped
					err = inst.CommitFields(db, set.NewSet("state"))
					if err != nil {
						return
					}
				}
				continue
			}
			break
		case instance.Deleting:
			if !busy.Contains(inst.Id) {
				busyLock.Lock()
				busy.Add(inst.Id)
				busyLock.Unlock()
				go func(inst *instance.Instance) {
					defer func() {
						busyLock.Lock()
						busy.Remove(inst.Id)
						busyLock.Unlock()
					}()
					db := database.GetDatabase()
					defer db.Close()

					e := qemu.Destroy(db, inst.GetVm())
					if e != nil {
						logrus.WithFields(logrus.Fields{
							"error": e,
						}).Error("sync: Failed to power off instance")
						return
					}

					err = instance.Remove(db, inst.Id)
					if err != nil {
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

func vmRunner() {
	time.Sleep(1 * time.Second)

	for {
		time.Sleep(1 * time.Second)

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

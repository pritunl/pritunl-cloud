package sync

import (
	"github.com/Sirupsen/logrus"
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/virtualbox"
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

	virts, err := virtualbox.GetVms(db)
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

				e := virtualbox.Create(db, inst.GetVm())
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
		virt := virtsMap[idInf.(bson.ObjectId)]
		if !busy.Contains(virt.Id) {
			busyLock.Lock()
			busy.Add(virt.Id)
			busyLock.Unlock()
			go func(virt *vm.VirtualMachine) {
				defer func() {
					busyLock.Lock()
					busy.Remove(virt.Id)
					busyLock.Unlock()
				}()
				db := database.GetDatabase()
				defer db.Close()

				e := virtualbox.Destroy(db, virt)
				time.Sleep(3 * time.Second)
				if e != nil {
					logrus.WithFields(logrus.Fields{
						"error": e,
					}).Error("sync: Failed to destroy instance")
					return
				}
			}(virt)
		}
	}

	for _, inst := range instances {
		virt := virtsMap[inst.Id]
		if virt == nil {
			continue
		}

		switch inst.State {
		case instance.Running:
			if (virt.State == vm.PowerOff || virt.State == vm.Aborted) &&
				!busy.Contains(inst.Id) {

				busyLock.Lock()
				busy.Add(virt.Id)
				busyLock.Unlock()
				go func(virt *vm.VirtualMachine) {
					defer func() {
						busyLock.Lock()
						busy.Remove(virt.Id)
						busyLock.Unlock()
					}()
					db := database.GetDatabase()
					defer db.Close()

					e := virtualbox.PowerOn(db, virt)
					if e != nil {
						logrus.WithFields(logrus.Fields{
							"error": e,
						}).Error("sync: Failed to power on instance")
						return
					}

					time.Sleep(3 * time.Second)
				}(virt)
				continue
			}
			break
		case instance.Stopped:
			if virt.State == vm.Running && !busy.Contains(inst.Id) {
				busyLock.Lock()
				busy.Add(virt.Id)
				busyLock.Unlock()
				go func(virt *vm.VirtualMachine) {
					defer func() {
						busyLock.Lock()
						busy.Remove(virt.Id)
						busyLock.Unlock()
					}()
					db := database.GetDatabase()
					defer db.Close()

					e := virtualbox.PowerOff(db, virt)
					if e != nil {
						logrus.WithFields(logrus.Fields{
							"error": e,
						}).Error("sync: Failed to power off instance")
						return
					}

					time.Sleep(10 * time.Second)
				}(virt)
				continue
			}
			break
		case instance.Updating:
			if !busy.Contains(inst.Id) {
				if virt.State != vm.PowerOff {
					busyLock.Lock()
					busy.Add(virt.Id)
					busyLock.Unlock()
					go func(virt *vm.VirtualMachine) {
						defer func() {
							busyLock.Lock()
							busy.Remove(virt.Id)
							busyLock.Unlock()
						}()
						db := database.GetDatabase()
						defer db.Close()

						e := virtualbox.PowerOff(db, virt)
						if e != nil {
							logrus.WithFields(logrus.Fields{
								"error": e,
							}).Error("sync: Failed to power off instance")
							return
						}
					}(virt)
					continue
				}

				if inst.Changed(virt) {
					busyLock.Lock()
					busy.Add(virt.Id)
					busyLock.Unlock()

					logrus.WithFields(logrus.Fields{
						"id":             virt.Id.Hex(),
						"memory_old":     virt.Memory,
						"memory":         inst.Memory,
						"processors_old": virt.Processors,
						"processors":     inst.Processors,
					}).Info("virtualbox: Resizing virtual machine")

					go func(inst *instance.Instance, virt *vm.VirtualMachine) {
						defer func() {
							busyLock.Lock()
							busy.Remove(virt.Id)
							busyLock.Unlock()
						}()
						db := database.GetDatabase()
						defer db.Close()

						e := virtualbox.Update(db, inst.GetVm())
						if e != nil {
							logrus.WithFields(logrus.Fields{
								"error": e,
							}).Error("sync: Failed to power off instance")
							return
						}

						time.Sleep(5 * time.Second)

						inst.State = instance.Stopped
						err = inst.CommitFields(db, set.NewSet("status"))
						if err != nil {
							return
						}
					}(inst, virt)
				} else {
					inst.State = instance.Stopped
					err = inst.CommitFields(db, set.NewSet("status"))
					if err != nil {
						return
					}
				}
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

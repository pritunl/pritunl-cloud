package state

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

func update() (err error) {
	db := database.GetDatabase()
	defer db.Close()

	virts, err := virtualbox.GetVms()
	if err != nil {
		return
	}

	curIds := set.NewSet()
	virtsMap := map[bson.ObjectId]*vm.VirtualMachine{}
	for _, virt := range virts {
		curIds.Add(virt.Id)
		virtsMap[virt.Id] = virt

		err = virt.Commit(db)
		if err != nil {
			return
		}
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
				e := virtualbox.Create(inst.GetVm())
				time.Sleep(5 * time.Second)
				if e != nil {
					logrus.WithFields(logrus.Fields{
						"error": e,
					}).Error("state: Failed to create instance")
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
				e := virtualbox.Destroy(virt)
				time.Sleep(3 * time.Second)
				if e != nil {
					logrus.WithFields(logrus.Fields{
						"error": e,
					}).Error("state: Failed to destroy instance")
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

		switch inst.Status {
		case instance.Running:
			if virt.State == "poweroff" && !busy.Contains(inst.Id) {
				busyLock.Lock()
				busy.Add(virt.Id)
				busyLock.Unlock()
				go func(virt *vm.VirtualMachine) {
					defer func() {
						busyLock.Lock()
						busy.Remove(virt.Id)
						busyLock.Unlock()
					}()
					e := virtualbox.PowerOn(virt)
					if e != nil {
						logrus.WithFields(logrus.Fields{
							"error": e,
						}).Error("state: Failed to power on instance")
						return
					}

					time.Sleep(3 * time.Second)
				}(virt)
				continue
			}
			break
		case instance.Stopped:
			if virt.State == "running" && !busy.Contains(inst.Id) {
				busyLock.Lock()
				busy.Add(virt.Id)
				busyLock.Unlock()
				go func(virt *vm.VirtualMachine) {
					defer func() {
						busyLock.Lock()
						busy.Remove(virt.Id)
						busyLock.Unlock()
					}()

					e := virtualbox.PowerOff(virt)
					if e != nil {
						logrus.WithFields(logrus.Fields{
							"error": e,
						}).Error("state: Failed to power off instance")
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

						e := virtualbox.PowerOff(virt)
						if e != nil {
							logrus.WithFields(logrus.Fields{
								"error": e,
							}).Error("state: Failed to power off instance")
							return
						}

						time.Sleep(10 * time.Second)
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
							}).Error("state: Failed to power off instance")
							return
						}

						time.Sleep(5 * time.Second)

						inst.Status = instance.Stopped
						err = inst.CommitFields(db, set.NewSet("status"))
						if err != nil {
							return
						}
					}(inst, virt)
				} else {
					inst.Status = instance.Stopped
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

func runner() (err error) {
	for {
		time.Sleep(1 * time.Second)

		err = update()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("state: Failed to update")
			continue
		}
	}
}

func Init() {
	go func() {
		err := runner()
		if err != nil {
			panic(err)
		}
	}()
}

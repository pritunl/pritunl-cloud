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
	creating       = set.NewSet()
	creatingLock   = sync.Mutex{}
	destroying     = set.NewSet()
	destroyingLock = sync.Mutex{}
)

func update() (err error) {
	db := database.GetDatabase()
	defer db.Close()

	coll := db.Instances()

	virts, err := virtualbox.GetVms()
	if err != nil {
		return
	}

	curIds := set.NewSet()
	virtsMap := map[bson.ObjectId]*vm.VirtualMachine{}
	for _, virt := range virts {
		curIds.Add(virt.Id)
		virtsMap[virt.Id] = virt

		addr := ""
		if len(virt.NetworkAdapters) > 0 {
			addr = virt.NetworkAdapters[0].IpAddress
		}

		err = coll.UpdateId(virt.Id, &bson.M{
			"$set": &bson.M{
				"status":    virt.State,
				"public_ip": addr,
			},
		})
		if err != nil {
			err = database.ParseError(err)
			if _, ok := err.(*database.NotFoundError); ok {
				err = nil
			} else {
				return
			}
		}
	}

	instances, err := instance.GetAll(db, &bson.M{
		"node": node.Self.Id,
	})

	newIds := set.NewSet()
	for _, inst := range instances {
		newIds.Add(inst.Id)
		if !curIds.Contains(inst.Id) && !creating.Contains(inst.Id) {
			go func(inst *instance.Instance) {
				creatingLock.Lock()
				creating.Add(inst.Id)
				creatingLock.Unlock()
				defer func() {
					creatingLock.Lock()
					creating.Remove(inst.Id)
					creatingLock.Unlock()
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
		if !destroying.Contains(virt.Id) {
			go func(virt *vm.VirtualMachine) {
				destroyingLock.Lock()
				destroying.Add(virt.Id)
				destroyingLock.Unlock()
				defer func() {
					destroyingLock.Lock()
					destroying.Remove(virt.Id)
					destroyingLock.Unlock()
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

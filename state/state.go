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
	"time"
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
	}

	instances, err := instance.GetAll(db, &bson.M{
		"node": node.Self.Id,
	})

	newIds := set.NewSet()
	for _, inst := range instances {
		newIds.Add(inst.Id)
		if !curIds.Contains(inst.Id) {
			err = virtualbox.Create(inst.GetVm())
			if err != nil {
				return
			}
		}
	}

	curIds.Subtract(newIds)
	for idInf := range curIds.Iter() {
		virt := virtsMap[idInf.(bson.ObjectId)]
		err = virtualbox.Destroy(virt)
		if err != nil {
			return
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

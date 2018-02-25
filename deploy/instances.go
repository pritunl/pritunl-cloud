package deploy

import (
	"github.com/Sirupsen/logrus"
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/qemu"
	"github.com/pritunl/pritunl-cloud/qms"
	"github.com/pritunl/pritunl-cloud/state"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"time"
)

var (
	instancesLock = utils.NewMultiTimeoutLock(3 * time.Minute)
)

type Instances struct {
	stat *state.State
}

func (s *Instances) create(inst *instance.Instance) {
	if instancesLock.Locked(inst.Id.Hex()) {
		return
	}

	lockId := instancesLock.Lock(inst.Id.Hex())
	go func() {
		defer func() {
			time.Sleep(3 * time.Second)
			instancesLock.Unlock(inst.Id.Hex(), lockId)
		}()

		db := database.GetDatabase()
		defer db.Close()

		err := qemu.Create(db, inst, inst.Virt)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("deploy: Failed to create instance")
			return
		}

		event.PublishDispatch(db, "instance.change")
	}()
}

func (s *Instances) start(inst *instance.Instance) {
	if instancesLock.Locked(inst.Id.Hex()) {
		return
	}

	lockId := instancesLock.Lock(inst.Id.Hex())
	go func() {
		defer func() {
			time.Sleep(3 * time.Second)
			instancesLock.Unlock(inst.Id.Hex(), lockId)
		}()

		db := database.GetDatabase()
		defer db.Close()

		err := qemu.PowerOn(db, inst.Virt)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("deploy: Failed to start instance")
			return
		}

		event.PublishDispatch(db, "instance.change")
	}()
}

func (s *Instances) stop(inst *instance.Instance) {
	if instancesLock.Locked(inst.Id.Hex()) {
		return
	}

	lockId := instancesLock.Lock(inst.Id.Hex())
	go func() {
		defer func() {
			time.Sleep(3 * time.Second)
			instancesLock.Unlock(inst.Id.Hex(), lockId)
		}()

		db := database.GetDatabase()
		defer db.Close()

		err := qemu.PowerOff(db, inst.Virt)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("deploy: Failed to stop instance")
			return
		}

		event.PublishDispatch(db, "instance.change")
	}()
}

func (s *Instances) restart(inst *instance.Instance) {
	if instancesLock.Locked(inst.Id.Hex()) {
		return
	}

	lockId := instancesLock.Lock(inst.Id.Hex())
	go func() {
		defer func() {
			time.Sleep(3 * time.Second)
			instancesLock.Unlock(inst.Id.Hex(), lockId)
		}()

		db := database.GetDatabase()
		defer db.Close()

		err := qemu.PowerOn(db, inst.Virt)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("deploy: Failed to restart instance")
			return
		}

		time.Sleep(1 * time.Second)

		err = qemu.PowerOff(db, inst.Virt)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("deploy: Failed to restart instance")
			return
		}

		inst.State = instance.Start
		err = inst.CommitFields(db, set.NewSet("state"))
		if err != nil {
			return
		}

		event.PublishDispatch(db, "instance.change")
	}()
}

func (s *Instances) destroy(inst *instance.Instance) {
	if instancesLock.Locked(inst.Id.Hex()) {
		return
	}

	lockId := instancesLock.Lock(inst.Id.Hex())
	go func() {
		defer func() {
			time.Sleep(3 * time.Second)
			instancesLock.Unlock(inst.Id.Hex(), lockId)
		}()

		db := database.GetDatabase()
		defer db.Close()

		err := qemu.Destroy(db, inst.Virt)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("deploy: Failed to power off instance")
			return
		}

		err = instance.Remove(db, inst.Id)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("deploy: Failed to remove instance")
			return
		}

		event.PublishDispatch(db, "instance.change")
		event.PublishDispatch(db, "disk.change")
	}()
}

func (s *Instances) diskRemove(inst *instance.Instance, remDisks []*vm.Disk) {
	if instancesLock.Locked(inst.Id.Hex()) {
		return
	}

	lockId := instancesLock.Lock(inst.Id.Hex())
	go func() {
		defer func() {
			time.Sleep(3 * time.Second)
			instancesLock.Unlock(inst.Id.Hex(), lockId)
		}()

		db := database.GetDatabase()
		defer db.Close()

		for _, dsk := range remDisks {
			e := qms.RemoveDisk(inst.Id, dsk)
			if e != nil {
				logrus.WithFields(logrus.Fields{
					"error": e,
				}).Error("sync: Failed to remove disk")
				return
			}
		}

		event.PublishDispatch(db, "instance.change")
		event.PublishDispatch(db, "disk.change")
	}()
}

func (s *Instances) diff(db *database.Database,
	inst *instance.Instance) (err error) {

	curVirt := s.stat.GetVirt(inst.Id)
	changed := inst.Changed(curVirt)
	addDisks, remDisks := inst.DiskChanged(curVirt)
	if len(addDisks) > 0 {
		changed = true
	}

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
		s.diskRemove(inst, remDisks)
	}

	return
}

func (s *Instances) Deploy() (err error) {
	instances := s.stat.Instances()

	db := database.GetDatabase()
	defer db.Close()

	for _, inst := range instances {
		curVirt := s.stat.GetVirt(inst.Id)

		if inst.State == instance.Destroy {
			s.destroy(inst)
			continue
		}

		if curVirt == nil {
			s.create(inst)
			continue
		}

		switch inst.State {
		case instance.Start:
			if curVirt.State == vm.Stopped || curVirt.State == vm.Failed {
				s.start(inst)
				continue
			}

			err = s.diff(db, inst)
			if err != nil {
				return
			}

			break
		case instance.Stop:
			if curVirt.State == vm.Running {
				s.stop(inst)
				continue
			}
			break
		case instance.Restart:
			if curVirt.State == vm.Running {
				s.restart(inst)
				continue
			}
			break
		}
	}

	return
}

func NewInstances(stat *state.State) *Instances {
	return &Instances{
		stat: stat,
	}
}

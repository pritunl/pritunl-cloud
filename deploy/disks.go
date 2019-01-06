package deploy

import (
	"github.com/Sirupsen/logrus"
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/pritunl-cloud/data"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/state"
	"github.com/pritunl/pritunl-cloud/utils"
	"time"
)

var (
	disksLock = utils.NewMultiTimeoutLock(5 * time.Minute)
)

type Disks struct {
	stat *state.State
}

func (d *Disks) provision(dsk *disk.Disk) {
	acquired, lockId := disksLock.LockOpen(dsk.Id.Hex())
	if !acquired {
		return
	}

	go func() {
		defer disksLock.Unlock(dsk.Id.Hex(), lockId)

		db := database.GetDatabase()
		defer db.Close()

		backingImage, err := data.CreateDisk(db, dsk)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("deploy: Failed to provision disk")
			return
		}

		dsk.State = disk.Available
		dsk.BackingImage = backingImage

		err = dsk.CommitFields(db, set.NewSet("state", "backing_image"))
		if err != nil {
			return
		}

		event.PublishDispatch(db, "disk.change")
	}()
}

func (d *Disks) snapshot(dsk *disk.Disk) {
	acquired, lockId := disksLock.LockOpen(dsk.Id.Hex())
	if !acquired {
		return
	}

	go func() {
		defer disksLock.Unlock(dsk.Id.Hex(), lockId)

		db := database.GetDatabase()
		defer db.Close()

		err := data.CreateSnapshot(db, dsk)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("deploy: Failed to snapshot disk")
		}

		dsk.State = disk.Available
		err = dsk.CommitFields(db, set.NewSet("state"))
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("deploy: Failed update disk state")
			time.Sleep(5 * time.Second)
			return
		}

		event.PublishDispatch(db, "disk.change")
	}()
}

func (d *Disks) destroy(dsk *disk.Disk) {
	if dsk.DeleteProtection {
		db := database.GetDatabase()
		defer db.Close()

		logrus.WithFields(logrus.Fields{
			"disk_id": dsk.Id.Hex(),
		}).Info("deploy: Delete protection ignore disk destroy")

		dsk.State = disk.Available
		dsk.CommitFields(db, set.NewSet("state"))

		event.PublishDispatch(db, "disk.change")

		return
	}

	if d.stat.DiskInUse(dsk.Instance, dsk.Id) {
		return
	}

	acquired, lockId := disksLock.LockOpen(dsk.Id.Hex())
	if !acquired {
		return
	}

	go func() {
		defer disksLock.Unlock(dsk.Id.Hex(), lockId)

		db := database.GetDatabase()
		defer db.Close()

		err := dsk.Destroy(db)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("deploy: Failed to destroy disk")
			time.Sleep(5 * time.Second)
			return
		}

		event.PublishDispatch(db, "disk.change")
	}()
}

func (d *Disks) Deploy() (err error) {
	disks := d.stat.Disks()

	for _, dsk := range disks {
		switch dsk.State {
		case disk.Provision:
			d.provision(dsk)
			break
		case disk.Snapshot:
			d.snapshot(dsk)
			break
		case disk.Destroy:
			d.destroy(dsk)
			break
		}
	}

	return
}

func NewDisks(stat *state.State) *Disks {
	return &Disks{
		stat: stat,
	}
}

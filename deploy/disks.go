package deploy

import (
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/pritunl-cloud/data"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/state"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/sirupsen/logrus"
)

var (
	disksLock     = utils.NewMultiTimeoutLock(5 * time.Minute)
	backupLimiter = utils.NewLimiter(3)
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

		virt := d.stat.GetVirt(dsk.Instance)
		err := data.CreateSnapshot(db, dsk, virt)
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

func (d *Disks) expand(dsk *disk.Disk) {
	acquired, lockId := disksLock.LockOpen(dsk.Id.Hex())
	if !acquired {
		return
	}

	go func() {
		defer disksLock.Unlock(dsk.Id.Hex(), lockId)

		db := database.GetDatabase()
		defer db.Close()

		inst := d.stat.GetInstace(dsk.Instance)
		if inst != nil {
			if inst.State != instance.Stop {
				inst.State = instance.Stop

				logrus.WithFields(logrus.Fields{
					"instance_id": inst.Id.Hex(),
					"disk_id":     dsk.Id.Hex(),
				}).Info("deploy: Stopping instance for resize")

				err := inst.CommitFields(db, set.NewSet("state"))
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"error": err,
					}).Error("deploy: Failed to commit instance state")
					return
				}

				return
			}

			virt := d.stat.GetVirt(inst.Id)
			if virt != nil && virt.State != vm.Stopped &&
				virt.State != vm.Failed {

				return
			}
		}

		err := data.ExpandDisk(db, dsk)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("deploy: Failed to expand disk")
		}

		dsk.State = disk.Available
		dsk.NewSize = 0
		err = dsk.CommitFields(db, set.NewSet("state", "size", "new_size"))
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

func (d *Disks) backup(dsk *disk.Disk) {
	if !backupLimiter.Acquire() {
		return
	}

	acquired, lockId := disksLock.LockOpen(dsk.Id.Hex())
	if !acquired {
		backupLimiter.Release()
		return
	}

	go func() {
		defer func() {
			time.Sleep(1 * time.Second)
			disksLock.Unlock(dsk.Id.Hex(), lockId)
			backupLimiter.Release()
		}()

		db := database.GetDatabase()
		defer db.Close()

		virt := d.stat.GetVirt(dsk.Instance)
		err := data.CreateBackup(db, dsk, virt)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("deploy: Failed to backup disk")
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
		event.PublishDispatch(db, "image.change")
	}()
}

func (d *Disks) restore(dsk *disk.Disk) {
	if !backupLimiter.Acquire() {
		return
	}

	acquired, lockId := disksLock.LockOpen(dsk.Id.Hex())
	if !acquired {
		backupLimiter.Release()
		return
	}

	go func() {
		defer func() {
			time.Sleep(1 * time.Second)
			disksLock.Unlock(dsk.Id.Hex(), lockId)
			backupLimiter.Release()
		}()

		db := database.GetDatabase()
		defer db.Close()

		inst := d.stat.GetInstace(dsk.Instance)
		if inst != nil {
			if inst.State != instance.Stop {
				inst.State = instance.Stop

				logrus.WithFields(logrus.Fields{
					"instance_id": inst.Id.Hex(),
					"disk_id":     dsk.Id.Hex(),
				}).Info("deploy: Stopping instance for restore")

				err := inst.CommitFields(db, set.NewSet("state"))
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"error": err,
					}).Error("deploy: Failed to commit instance state")
					return
				}

				return
			}

			virt := d.stat.GetVirt(inst.Id)
			if virt != nil && virt.State != vm.Stopped &&
				virt.State != vm.Failed {

				return
			}
		}

		err := data.RestoreBackup(db, dsk)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"instance_id": inst.Id.Hex(),
				"disk_id":     dsk.Id.Hex(),
				"error":       err,
			}).Error("deploy: Failed to restore disk")
		}

		dsk.State = disk.Available
		err = dsk.CommitFields(db, set.NewSet("state"))
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"instance_id": inst.Id.Hex(),
				"disk_id":     dsk.Id.Hex(),
				"error":       err,
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

func (d *Disks) scheduleBackup(dsk *disk.Disk) {
	if time.Since(dsk.LastBackup) < 24*time.Hour {
		return
	}

	if !backupLimiter.Acquire() {
		return
	}

	acquired, lockId := disksLock.LockOpen(dsk.Id.Hex())
	if !acquired {
		backupLimiter.Release()
		return
	}

	go func() {
		defer func() {
			time.Sleep(1 * time.Second)
			disksLock.Unlock(dsk.Id.Hex(), lockId)
			backupLimiter.Release()
		}()

		db := database.GetDatabase()
		defer db.Close()

		if dsk.State != disk.Available {
			return
		}

		logrus.WithFields(logrus.Fields{
			"disk_id": dsk.Id.Hex(),
		}).Info("deploy: Scheduling automatic disk backup")

		dsk.State = disk.Backup
		dsk.LastBackup = time.Now()
		err := dsk.CommitFields(db, set.NewSet("state", "last_backup"))
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("deploy: Failed update disk state")
			time.Sleep(5 * time.Second)
			return
		}

		event.PublishDispatch(db, "disk.change")

		virt := d.stat.GetVirt(dsk.Instance)
		err = data.CreateBackup(db, dsk, virt)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("deploy: Failed to backup disk")
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
		event.PublishDispatch(db, "image.change")
	}()
}

func (d *Disks) Deploy() (err error) {
	disks := d.stat.Disks()

	backupHour := settings.System.DiskBackupTime
	backupWindow := settings.System.DiskBackupWindow
	utcHour := time.Now().UTC().Hour()
	backupActive := false
	if utcHour >= backupHour && utcHour <= (backupHour+backupWindow) {
		backupActive = true
	}

	for _, dsk := range disks {
		switch dsk.State {
		case disk.Provision:
			d.provision(dsk)
			break
		case disk.Snapshot:
			d.snapshot(dsk)
			break
		case disk.Backup:
			d.backup(dsk)
			break
		case disk.Restore:
			d.restore(dsk)
			break
		case disk.Destroy:
			d.destroy(dsk)
			break
		case disk.Available:
			if backupActive && dsk.Backup {
				d.scheduleBackup(dsk)
			}
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

package deploy

import (
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/constants"
	"github.com/pritunl/pritunl-cloud/data"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/errortypes"
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

		if constants.Interrupt {
			return
		}

		newSize, backingImage, err := data.CreateDisk(db, dsk)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("deploy: Failed to provision disk")
			return
		}

		fields := set.NewSet("state", "backing_image")

		dsk.State = disk.Available
		dsk.BackingImage = backingImage

		if newSize != 0 {
			fields.Add("size")
			dsk.Size = newSize
		}

		err = dsk.CommitFields(db, fields)
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

		if constants.Interrupt {
			return
		}

		if dsk.Type != disk.Qcow2 {
			logrus.WithFields(logrus.Fields{
				"disk_id":   dsk.Id.Hex(),
				"disk_type": dsk.Type,
			}).Error("deploy: Disk type does not support snapshot")
		} else {
			virt := d.stat.GetVirt(dsk.Instance)
			if virt == nil {
				err := &errortypes.ReadError{
					errors.New("deploy: Failed to load virt"),
				}
				logrus.WithFields(logrus.Fields{
					"disk_id": dsk.Id.Hex(),
					"error":   err,
				}).Error("deploy: Failed to load virt")
				return
			}

			err := data.CreateSnapshot(db, dsk, virt)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Error("deploy: Failed to snapshot disk")
			}
		}

		dsk.State = disk.Available
		err := dsk.CommitFields(db, set.NewSet("state"))
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"disk_id":   dsk.Id.Hex(),
				"disk_type": dsk.Type,
				"error":     err,
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

		if constants.Interrupt {
			return
		}

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

		if constants.Interrupt {
			return
		}

		if dsk.Type != disk.Qcow2 {
			logrus.WithFields(logrus.Fields{
				"disk_id":   dsk.Id.Hex(),
				"disk_type": dsk.Type,
			}).Error("deploy: Disk type does not support backup")
		} else {
			virt := d.stat.GetVirt(dsk.Instance)
			if virt == nil {
				err := &errortypes.ReadError{
					errors.New("deploy: Failed to load virt"),
				}
				logrus.WithFields(logrus.Fields{
					"disk_id": dsk.Id.Hex(),
					"error":   err,
				}).Error("deploy: Failed to load virt")
				return
			}

			err := data.CreateBackup(db, dsk, virt)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Error("deploy: Failed to backup disk")
			}
		}

		dsk.State = disk.Available
		err := dsk.CommitFields(db, set.NewSet("state"))
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

		if constants.Interrupt {
			return
		}

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

		if dsk.Type != disk.Qcow2 {
			logrus.WithFields(logrus.Fields{
				"disk_id":   dsk.Id.Hex(),
				"disk_type": dsk.Type,
			}).Error("deploy: Disk type does not support restore")
		} else {
			err := data.RestoreBackup(db, dsk)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"disk_id": dsk.Id.Hex(),
					"error":   err,
				}).Error("deploy: Failed to restore disk")
			}
		}

		dsk.State = disk.Available
		err := dsk.CommitFields(db, set.NewSet("state"))
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"disk_id": dsk.Id.Hex(),
				"error":   err,
			}).Error("deploy: Failed update disk state")
			time.Sleep(5 * time.Second)
			return
		}

		event.PublishDispatch(db, "disk.change")
	}()
}

func (d *Disks) destroy(db *database.Database, dsk *disk.Disk) {
	var inst *instance.Instance
	if !dsk.Instance.IsZero() {
		inst = d.stat.GetInstace(dsk.Instance)
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

		if constants.Interrupt {
			return
		}

		if dsk.DeleteProtection {
			logrus.WithFields(logrus.Fields{
				"disk_id": dsk.Id.Hex(),
			}).Info("deploy: Delete protection ignore disk destroy")

			dsk.State = disk.Available
			_ = dsk.CommitFields(db, set.NewSet("state"))

			event.PublishDispatch(db, "disk.change")

			return
		}

		if inst != nil && inst.DeleteProtection {
			logrus.WithFields(logrus.Fields{
				"disk_id": dsk.Id.Hex(),
			}).Info("deploy: Instance delete protection ignore disk destroy")

			dsk.State = disk.Available
			_ = dsk.CommitFields(db, set.NewSet("state"))

			event.PublishDispatch(db, "disk.change")

			return
		}

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

		if constants.Interrupt {
			return
		}

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
		if virt == nil {
			err := &errortypes.ReadError{
				errors.New("deploy: Failed to load virt"),
			}
			logrus.WithFields(logrus.Fields{
				"disk_id": dsk.Id.Hex(),
				"error":   err,
			}).Error("deploy: Failed to load virt")
			return
		}

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

func (d *Disks) Deploy(db *database.Database) (err error) {
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
		case disk.Expand:
			d.expand(dsk)
			break
		case disk.Destroy:
			d.destroy(db, dsk)
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

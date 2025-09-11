package backup

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/config"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/qmp"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/sirupsen/logrus"
)

type Backup struct {
	Destination string
	node        *node.Node
	virtPath    string
	errorCount  int
}

func (b *Backup) backupDisk(db *database.Database,
	dsk *disk.Disk, dest string) (err error) {

	online := false
	if !dsk.Instance.IsZero() {
		inst, e := instance.Get(db, dsk.Instance)
		if e != nil {
			err = e
			return
		}

		if inst != nil {
			if inst.State == vm.Starting {
				time.Sleep(5 * time.Second)
				online = true
			}

			if inst.State == vm.Running {
				online = true
			}
		}
	}

	_ = os.Remove(dest)

	if online {
		err = qmp.BackupDisk(dsk.Instance, dsk, dest)
		if err != nil {
			if _, ok := err.(*qmp.DiskNotFound); ok {
				online = false
				err = nil
			} else {
				return
			}
		}
	}

	if !online {
		dskPth := path.Join(b.virtPath, "disks",
			fmt.Sprintf("%s.qcow2", dsk.Id.Hex()))
		err = utils.Exec("", "cp", dskPth, dest)
		if err != nil {
			return
		}
	}

	return
}

func (b *Backup) backupDisks(db *database.Database) (err error) {
	logrus.WithFields(logrus.Fields{
		"node_id": b.node.Id.Hex(),
	}).Info("backup: Exporting disks")

	trashDir := path.Join(b.Destination, "trash")
	disksDir := path.Join(b.Destination, "disks")
	err = utils.ExistsMkdir(disksDir, 0755)
	if err != nil {
		return
	}

	disks, err := disk.GetAll(db, &bson.M{
		"node": b.node.Id,
	})

	diskFilenames := set.NewSet()
	for _, dsk := range disks {
		filename := fmt.Sprintf("%s.qcow2", dsk.Id.Hex())
		diskFilenames.Add(filename)

		destPath := path.Join(disksDir, filename)

		err = b.backupDisk(db, dsk, destPath)
		if err != nil {
			b.errorCount += 1
			logrus.WithFields(logrus.Fields{
				"disk_id": dsk.Id.Hex(),
				"error":   err,
			}).Error("qemu: Failed to backup disk")
		} else {
			logrus.WithFields(logrus.Fields{
				"node_id": b.node.Id.Hex(),
				"disk_id": dsk.Id.Hex(),
			}).Info("backup: Disk exported")
		}
	}

	exportedDisks, err := ioutil.ReadDir(disksDir)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "backup: Failed to read disks directory"),
		}
		return
	}

	trashDiskDir := path.Join(trashDir, "disks")
	for _, item := range exportedDisks {
		filename := item.Name()
		diskPath := path.Join(disksDir, filename)
		newDiskPath := path.Join(trashDiskDir, filename)
		idStr := strings.Split(path.Base(filename), ".")[0]

		if !diskFilenames.Contains(filename) {
			err = utils.ExistsMkdir(trashDiskDir, 0755)
			if err != nil {
				return
			}

			err = os.Rename(diskPath, newDiskPath)
			if err != nil {
				b.errorCount += 1
				logrus.WithFields(logrus.Fields{
					"node_id": b.node.Id.Hex(),
					"disk_id": idStr,
					"error":   err,
				}).Error("backup: Failed to move disk to trash")
			} else {
				logrus.WithFields(logrus.Fields{
					"node_id": b.node.Id.Hex(),
					"disk_id": idStr,
				}).Info("backup: Disk moved to trash")
			}
		}
	}

	return
}

func (b *Backup) backupBackingDisks(db *database.Database) (err error) {
	logrus.WithFields(logrus.Fields{
		"node_id": b.node.Id.Hex(),
	}).Info("backup: Exporting backing disks")

	trashDir := path.Join(b.Destination, "trash")
	backingDisksDir := path.Join(b.Destination, "backing")
	curBackingDisksDir := path.Join(b.virtPath, "backing")
	err = utils.ExistsMkdir(backingDisksDir, 0755)
	if err != nil {
		return
	}

	exists, err := utils.Exists(curBackingDisksDir)
	if err != nil {
		return
	}

	curBackingDisks := []os.FileInfo{}
	if exists {
		curBackingDisks, err = ioutil.ReadDir(curBackingDisksDir)
		if err != nil {
			err = &errortypes.ReadError{
				errors.Wrap(
					err,
					"backup: Failed to read backing disks directory",
				),
			}
			return
		}
	}

	backingFilenames := set.NewSet()
	for _, item := range curBackingDisks {
		filename := item.Name()
		backingFilenames.Add(filename)

		backingPath := path.Join(curBackingDisksDir, filename)
		newBackingPath := path.Join(backingDisksDir, filename)

		_ = os.Remove(newBackingPath)

		err = utils.Exec("", "cp", backingPath, newBackingPath)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"node_id":      b.node.Id.Hex(),
			"backing_disk": filename,
		}).Info("backup: Backing disk exported")
	}

	exportedBackingDisks, err := ioutil.ReadDir(backingDisksDir)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(
				err,
				"backup: Failed to read backing disks directory",
			),
		}
		return
	}

	trashBackingDir := path.Join(trashDir, "backing")
	for _, item := range exportedBackingDisks {
		filename := item.Name()
		backingPath := path.Join(backingDisksDir, filename)
		newBackingPath := path.Join(trashBackingDir, filename)

		if !backingFilenames.Contains(filename) {
			err = utils.ExistsMkdir(trashBackingDir, 0755)
			if err != nil {
				return
			}

			err = os.Rename(backingPath, newBackingPath)
			if err != nil {
				b.errorCount += 1
				logrus.WithFields(logrus.Fields{
					"node_id":      b.node.Id.Hex(),
					"backing_disk": filename,
					"error":        err,
				}).Error("backup: Failed to move backing disk to trash")
			} else {
				logrus.WithFields(logrus.Fields{
					"node_id":      b.node.Id.Hex(),
					"backing_disk": filename,
				}).Info("backup: Backing disk moved to trash")
			}
		}
	}

	return
}

func (b *Backup) Run() (err error) {
	db := database.GetDatabase()
	defer db.Close()

	ndeId, err := bson.ObjectIDFromHex(config.Config.NodeId)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "backup: Failed to parse ObjectId"),
		}
		return
	}

	nde, err := node.Get(db, ndeId)
	if err != nil {
		return
	}

	b.node = nde
	b.virtPath = nde.GetVirtPath()

	err = b.backupDisks(db)
	if err != nil {
		return
	}

	err = b.backupBackingDisks(db)
	if err != nil {
		return
	}

	if b.errorCount > 0 {
		err = &errortypes.ExecError{
			errors.Wrap(err, "backup: Backup encountered errors"),
		}
		return
	}

	return
}

func New(dest string) *Backup {
	return &Backup{
		Destination: dest,
	}
}

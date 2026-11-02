package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/config"
	"github.com/pritunl/pritunl-cloud/data"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/guestfs"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/pool"
	"github.com/sirupsen/logrus"
)

func ScanDisk(diskIdStr string) (err error) {
	guestfs.Debug = true

	db := database.GetDatabase()
	defer db.Close()

	ndeId, err := bson.ObjectIDFromHex(config.Config.NodeId)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "cmd: Failed to parse node ObjectId"),
		}
		return
	}

	diskId, err := bson.ObjectIDFromHex(diskIdStr)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "cmd: Failed to parse disk ObjectId"),
		}
		return
	}

	nde, err := node.Get(db, ndeId)
	if err != nil {
		return
	}

	node.Self = nde

	dsk, err := disk.Get(db, diskId)
	if err != nil {
		return
	}

	if dsk.Type != disk.Lvm && dsk.Node != nde.Id {
		err = &errortypes.ReadError{
			errors.New("cmd: Disk is not on this node"),
		}
		return
	}

	inUse, err := data.DiskInUse(db, dsk)
	if err != nil {
		return
	}

	if inUse {
		err = &errortypes.WriteError{
			errors.New("cmd: Disk is attached to a running instance, " +
				"stop the instance before scanning"),
		}
		return
	}

	fields := logrus.Fields{
		"node_id":     nde.Id.Hex(),
		"node_name":   nde.Name,
		"disk_id":     dsk.Id.Hex(),
		"disk_name":   dsk.Name,
		"disk_type":   dsk.Type,
		"disk_node":   dsk.Node.Hex(),
		"instance":    dsk.Instance.Hex(),
		"file_system": dsk.FileSystem,
	}

	if dsk.Type == disk.Lvm {
		pl, e := pool.Get(db, dsk.Pool)
		if e != nil {
			err = e
			return
		}

		fields["pool"] = pl.Name
		fields["vg_name"] = pl.VgName
		fields["lv_name"] = dsk.Id.Hex()
	} else {
		dskPath, e := dsk.GetPath(db)
		if e != nil {
			err = e
			return
		}

		fields["path"] = dskPath
	}

	logrus.WithFields(fields).Info("cmd: Scanning disk file system")

	start := time.Now()

	err = data.ScanDisk(db, dsk)
	if err != nil {
		return
	}

	partitions := 0
	if dsk.Metadata != nil {
		partitions = len(dsk.Metadata.Partitions)
	}

	logrus.WithFields(logrus.Fields{
		"disk_id":    dsk.Id.Hex(),
		"partitions": partitions,
		"elapsed":    time.Since(start).Round(time.Millisecond).String(),
	}).Info("cmd: Disk scan complete")

	err = dsk.CommitFields(db, set.NewSet("metadata"))
	if err != nil {
		return
	}

	_ = event.PublishDispatch(db, "disk.change")

	output, err := json.MarshalIndent(dsk.Metadata, "", "  ")
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "cmd: Failed to marshal disk metadata"),
		}
		return
	}

	fmt.Println(string(output))

	return
}

package qmp

import (
	"path"
	"strings"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/sirupsen/logrus"
)

type driveBackupArgs struct {
	Device string `json:"device"`
	Sync   string `json:"sync"`
	Target string `json:"target"`
	Format string `json:"format"`
}

type blockDeviceImage struct {
	Filename string `json:"filename"`
}

type blockDeviceInserted struct {
	Image blockDeviceImage `json:"image"`
}

type blockDevice struct {
	Device   string              `json:"device"`
	Inserted blockDeviceInserted `json:"inserted"`
}

type blockDeviceReturn struct {
	Return []*blockDevice `json:"return"`
	Error  *CommandError  `json:"error"`
}

func driveGetDevice(vmId bson.ObjectID, dsk *disk.Disk) (
	name string, err error) {

	cmd := &Command{
		Execute: "query-block",
	}

	returnData := &blockDeviceReturn{}
	err = RunCommand(vmId, cmd, returnData)
	if err != nil {
		return
	}

	if returnData.Error != nil {
		err = &errortypes.ApiError{
			errors.Newf("qmp: Return error %s", returnData.Error.Desc),
		}
		return
	}

	if returnData.Return == nil {
		err = &errortypes.ParseError{
			errors.Newf("qmp: Return nil"),
		}
		return
	}

	for _, blockDev := range returnData.Return {
		idStr := strings.Split(path.Base(
			blockDev.Inserted.Image.Filename), ".")[0]

		diskId, err := bson.ObjectIDFromHex(idStr)
		if err != nil {
			continue
		}

		if diskId == dsk.Id {
			name = blockDev.Device
			break
		}
	}

	return
}

func driveBackup(vmId bson.ObjectID, dsk *disk.Disk,
	destPth string) (deviceName string, err error) {

	deviceName, err = driveGetDevice(vmId, dsk)
	if err != nil {
		return
	}

	if deviceName == "" {
		err = &DiskNotFound{
			errors.Newf("qmp: Disk not found %s", dsk.Id.Hex()),
		}
		return
	}

	cmd := &Command{
		Execute: "drive-backup",
		Arguments: &driveBackupArgs{
			Device: deviceName,
			Sync:   "full",
			Target: destPth,
			Format: "qcow2",
		},
	}

	returnData := &CommandReturn{}
	err = RunCommand(vmId, cmd, returnData)
	if err != nil {
		return
	}

	if returnData.Error != nil {
		err = &errortypes.ApiError{
			errors.Newf("qmp: Return error %s", returnData.Error.Desc),
		}
		return
	}

	time.Sleep(1 * time.Second)

	return
}

func driveBackupCheck(vmId bson.ObjectID, deviceName string) (
	complete bool, err error) {

	cmd := &Command{
		Execute: "query-jobs",
	}

	returnData := &JobStatusReturn{}
	err = RunCommand(vmId, cmd, returnData)
	if err != nil {
		return
	}

	if returnData.Error != nil {
		err = &errortypes.ApiError{
			errors.Newf("qmp: Return error %s", returnData.Error.Desc),
		}
		return
	}

	if returnData.Return == nil {
		err = &errortypes.ParseError{
			errors.Newf("qmp: Return nil"),
		}
		return
	}

	for _, status := range returnData.Return {
		if status.Type == "backup" &&
			status.Id == deviceName &&
			status.Status != "concluded" {

			return
		}
	}

	complete = true

	return
}

func BackupDisk(vmId bson.ObjectID, dsk *disk.Disk,
	destPth string) (err error) {

	logrus.WithFields(logrus.Fields{
		"instance_id": vmId.Hex(),
		"disk_id":     dsk.Id.Hex(),
	}).Info("qmp: Backing up disk")

	deviceName, err := driveBackup(vmId, dsk, destPth)
	if err != nil {
		return
	}

	for {
		complete, e := driveBackupCheck(vmId, deviceName)
		if e != nil {
			err = e
			return
		}

		if complete {
			break
		}

		time.Sleep(3 * time.Second)
	}

	return
}

package qmp

import (
	"bytes"
	"encoding/json"
	"net"
	"path"
	"strings"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/sirupsen/logrus"
)

type cmdBase struct {
	Execute   string      `json:"execute"`
	Arguments interface{} `json:"arguments,omitempty"`
}

type driveBackupArgs struct {
	Device string `json:"device"`
	Sync   string `json:"sync"`
	Target string `json:"target"`
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
	Error  *cmdError      `json:"error"`
}

type jobStatus struct {
	Id     string `json:"id"`
	Type   string `json:"type"`
	Status string `json:"status"`
}

type jobStatusReturn struct {
	Return []*jobStatus `json:"return"`
	Error  *cmdError    `json:"error"`
}

type cmdError struct {
	Class string `json:"class"`
	Desc  string `json:"desc"`
}

type cmdReturn struct {
	Return interface{} `json:"return"`
	Error  *cmdError   `json:"error"`
}

var (
	socketsLock = utils.NewMultiTimeoutLock(1 * time.Minute)
)

func runCommand(vmId primitive.ObjectID, cmd *cmdBase,
	cmdReturn interface{}) (err error) {

	sockPath := paths.GetQmpSockPath(vmId)

	lockId := socketsLock.Lock(vmId.Hex())
	defer socketsLock.Unlock(vmId.Hex(), lockId)

	conn, err := net.DialTimeout(
		"unix",
		sockPath,
		1*time.Second,
	)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "qmp: Failed to open socket"),
		}
		return
	}
	defer conn.Close()

	err = conn.SetDeadline(time.Now().Add(10 * time.Second))
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "qmp: Failed set deadline"),
		}
		return
	}

	initCmd := &cmdBase{
		Execute: "qmp_capabilities",
	}

	cmdData, err := json.Marshal(initCmd)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "qmp: Failed to marshal command"),
		}
		return
	}

	_, err = conn.Write([]byte(string(cmdData) + "\n"))
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "qmp: Failed to write socket"),
		}
		return
	}

	time.Sleep(100 * time.Millisecond)

	cmdData, err = json.Marshal(cmd)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "qmp: Failed to marshal command"),
		}
		return
	}

	_, err = conn.Write([]byte(string(cmdData) + "\n"))
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "qmp: Failed to write socket"),
		}
		return
	}

	buffer := make([]byte, 100000)
	for {
		buf := make([]byte, 10000)
		n, e := conn.Read(buf)
		if e != nil {
			err = &errortypes.ReadError{
				errors.Wrap(e, "qmp: Failed to read socket"),
			}
			return
		}
		buffer = append(buffer, buf[:n]...)

		if bytes.Count(bytes.TrimSpace(buffer), []byte(`"return"`)) > 1 ||
			bytes.Contains(bytes.TrimSpace(buffer), []byte(`"error"`)) {

			break
		}
	}

	initReturn := false
	returnDataStr := ""
	lines := strings.Split(string(buffer), "\n")
	for _, line := range lines {
		if strings.Contains(line, `"return"`) {
			if initReturn {
				returnDataStr = line
				break
			} else {
				initReturn = true
				continue
			}
		} else if strings.Contains(line, `"return"`) ||
			strings.Contains(line, `"error"`) {

			returnDataStr = line
			break
		}
	}

	err = json.Unmarshal([]byte(returnDataStr), cmdReturn)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrapf(
				err,
				"qmp: Failed to unmarshal return %s",
				returnDataStr,
			),
		}
		return
	}

	return
}

func driveGetDevice(vmId primitive.ObjectID, dsk *disk.Disk) (
	name string, err error) {

	cmd := &cmdBase{
		Execute: "query-block",
	}

	returnData := &blockDeviceReturn{}
	err = runCommand(vmId, cmd, returnData)
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

		diskId, err := primitive.ObjectIDFromHex(idStr)
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

func driveBackup(vmId primitive.ObjectID, dsk *disk.Disk,
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

	cmd := &cmdBase{
		Execute: "drive-backup",
		Arguments: &driveBackupArgs{
			Device: deviceName,
			Sync:   "full",
			Target: destPth,
		},
	}

	returnData := &cmdReturn{}
	err = runCommand(vmId, cmd, returnData)
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

func driveBackupCheck(vmId primitive.ObjectID, deviceName string) (
	complete bool, err error) {

	cmd := &cmdBase{
		Execute: "query-jobs",
	}

	returnData := &jobStatusReturn{}
	err = runCommand(vmId, cmd, returnData)
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

func BackupDisk(vmId primitive.ObjectID, dsk *disk.Disk,
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

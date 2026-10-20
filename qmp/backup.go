package qmp

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/sirupsen/logrus"
)

const (
	backupPollInterval   = 3 * time.Second
	backupLogInterval    = 60 * time.Second
	backupCancelTimeout  = 2 * time.Minute
	backupCancelInterval = 1 * time.Second
)

type namedNode struct {
	NodeName string `json:"node-name"`
	File     string `json:"file"`
}

type namedNodesReturn struct {
	Return []namedNode   `json:"return"`
	Error  *CommandError `json:"error"`
}

type backupTargetFile struct {
	Driver   string `json:"driver"`
	Filename string `json:"filename"`
}

type backupTargetArgs struct {
	Driver   string           `json:"driver"`
	NodeName string           `json:"node-name"`
	File     backupTargetFile `json:"file"`
}

type blockdevBackupArgs struct {
	JobId       string `json:"job-id"`
	Device      string `json:"device"`
	Target      string `json:"target"`
	Sync        string `json:"sync"`
	AutoDismiss bool   `json:"auto-dismiss"`
}

type BackupJob struct {
	vmId       bson.ObjectID
	dskId      bson.ObjectID
	jobId      string
	targetNode string
}

func backupJobId(dskId bson.ObjectID) string {
	return fmt.Sprintf("backup_%s", dskId.Hex())
}

func backupTargetNode(dskId bson.ObjectID) string {
	return fmt.Sprintf("bt_%s", dskId.Hex())
}

func apiError(cmdErr *CommandError) error {
	return &errortypes.ApiError{
		errors.Newf("qmp: Return error %s", cmdErr.Desc),
	}
}

func isNotFoundError(cmdErr *CommandError) bool {
	if cmdErr == nil {
		return false
	}

	desc := strings.ToLower(cmdErr.Desc)

	return strings.Contains(desc, "not found") ||
		strings.Contains(desc, "failed to find")
}

func (c *Connection) run(cmd *Command) (cmdErr *CommandError, err error) {
	returnData := &CommandReturn{}
	err = c.Send(cmd, returnData)
	if err != nil {
		return
	}

	cmdErr = returnData.Error

	return
}

func runChecked(vmId bson.ObjectID, cmd *Command) (
	cmdErr *CommandError, err error) {

	returnData := &CommandReturn{}
	err = RunCommand(vmId, cmd, returnData)
	if err != nil {
		return
	}

	cmdErr = returnData.Error

	return
}

func GetDiskNode(vmId, dskId bson.ObjectID) (
	device string, size int64, err error) {

	cmd := &Command{
		Execute: "query-block",
	}

	returnData := &blockQueryReturn{}
	err = RunCommand(vmId, cmd, returnData)
	if err != nil {
		return
	}

	if returnData.Error != nil {
		err = apiError(returnData.Error)
		return
	}

	if returnData.Return == nil {
		err = &errortypes.ParseError{
			errors.Newf("qmp: Return nil"),
		}
		return
	}

	dskHex := dskId.Hex()
	for _, blockDev := range returnData.Return {
		inserted := blockDev.Inserted
		nodeName := blockDev.diskNode()

		if blockDev.Device != "disk_"+dskHex &&
			nodeName != "fd_"+dskHex &&
			path.Base(inserted.Image.Filename) != dskHex+".qcow2" {

			continue
		}

		device = blockDev.Device
		if device == "" {
			device = nodeName
		}
		if device == "" {
			device = inserted.NodeName
		}
		size = inserted.Image.VirtualSize

		break
	}

	if device == "" {
		err = &DiskNotFound{
			errors.Newf("qmp: Disk not found %s", dskHex),
		}
		return
	}

	if size <= 0 {
		err = &errortypes.ParseError{
			errors.Newf("qmp: Disk %s has no virtual size", dskHex),
		}
		return
	}

	return
}

func (j *BackupJob) targetFile() (pth string, err error) {
	cmd := &Command{
		Execute: "query-named-block-nodes",
	}

	returnData := &namedNodesReturn{}
	err = RunCommand(j.vmId, cmd, returnData)
	if err != nil {
		return
	}

	if returnData.Error != nil {
		err = apiError(returnData.Error)
		return
	}

	for _, nde := range returnData.Return {
		if nde.NodeName == j.targetNode {
			pth = nde.File
			break
		}
	}

	return
}

func (j *BackupJob) cancelStale(status *JobStatus) (err error) {
	logrus.WithFields(logrus.Fields{
		"instance_id": j.vmId.Hex(),
		"disk_id":     j.dskId.Hex(),
		"status":      status.Status,
	}).Warning("qmp: Cancelling orphaned backup job")

	if status.Status != "aborting" {
		cmdErr, e := runChecked(j.vmId, &Command{
			Execute: "job-cancel",
			Arguments: &CommandId{
				Id: j.jobId,
			},
		})
		if e != nil {
			err = e
			return
		}

		if cmdErr != nil && !isNotFoundError(cmdErr) {
			logrus.WithFields(logrus.Fields{
				"instance_id": j.vmId.Hex(),
				"disk_id":     j.dskId.Hex(),
				"error":       cmdErr.Desc,
			}).Warning("qmp: Backup job cancel rejected")
		}
	}

	start := time.Now()
	for {
		time.Sleep(backupCancelInterval)

		status, err = j.find()
		if err != nil {
			return
		}

		if status == nil || status.Status == "concluded" {
			return
		}

		if time.Since(start) > backupCancelTimeout {
			err = &errortypes.TimeoutError{
				errors.Newf("qmp: Timeout cancelling backup job %s", j.jobId),
			}
			return
		}
	}
}

func (j *BackupJob) clearStale() (err error) {
	status, err := j.find()
	if err != nil {
		return
	}

	if status != nil && status.Status != "concluded" {
		err = j.cancelStale(status)
		if err != nil {
			return
		}

		status, err = j.find()
		if err != nil {
			return
		}
	}

	if status != nil {
		logrus.WithFields(logrus.Fields{
			"instance_id": j.vmId.Hex(),
			"disk_id":     j.dskId.Hex(),
			"job_error":   status.Error,
		}).Warning("qmp: Dismissing stale backup job")

		cmdErr, e := runChecked(j.vmId, &Command{
			Execute: "job-dismiss",
			Arguments: &CommandId{
				Id: j.jobId,
			},
		})
		if e != nil {
			err = e
			return
		}

		if cmdErr != nil && !isNotFoundError(cmdErr) {
			err = apiError(cmdErr)
			return
		}
	}

	targetPth, err := j.targetFile()
	if err != nil {
		return
	}

	cmdErr, err := runChecked(j.vmId, &Command{
		Execute: "blockdev-del",
		Arguments: &CommandNode{
			NodeName: j.targetNode,
		},
	})
	if err != nil {
		return
	}

	if cmdErr != nil {
		if !isNotFoundError(cmdErr) {
			err = apiError(cmdErr)
			return
		}

		return
	}

	logrus.WithFields(logrus.Fields{
		"instance_id": j.vmId.Hex(),
		"disk_id":     j.dskId.Hex(),
		"target_path": targetPth,
	}).Warning("qmp: Removed stale backup target")

	if targetPth != "" {
		e := os.Remove(targetPth)
		if e != nil && !os.IsNotExist(e) {
			logrus.WithFields(logrus.Fields{
				"instance_id": j.vmId.Hex(),
				"disk_id":     j.dskId.Hex(),
				"target_path": targetPth,
				"error":       e,
			}).Error("qmp: Failed to remove stale backup target")
		}
	}

	return
}

func BackupDisk(vmId, dskId bson.ObjectID, device, targetPth string) (
	job *BackupJob, err error) {

	job = &BackupJob{
		vmId:       vmId,
		dskId:      dskId,
		jobId:      backupJobId(dskId),
		targetNode: backupTargetNode(dskId),
	}

	logrus.WithFields(logrus.Fields{
		"instance_id": vmId.Hex(),
		"disk_id":     dskId.Hex(),
		"device":      device,
		"target_path": targetPth,
	}).Info("qmp: Starting disk backup")

	err = job.clearStale()
	if err != nil {
		job = nil
		return
	}

	conn := NewConnection(vmId, true)
	defer conn.Close()

	conn.SetDeadline(30 * time.Second)

	_, err = conn.Connect()
	if err != nil {
		job = nil
		return
	}

	cmdErr, err := conn.run(&Command{
		Execute: "blockdev-add",
		Arguments: &backupTargetArgs{
			Driver:   "qcow2",
			NodeName: job.targetNode,
			File: backupTargetFile{
				Driver:   "file",
				Filename: targetPth,
			},
		},
	})
	if err != nil {
		job = nil
		return
	}

	if cmdErr != nil {
		job = nil
		err = apiError(cmdErr)
		return
	}

	cmdErr, err = conn.run(&Command{
		Execute: "blockdev-backup",
		Arguments: &blockdevBackupArgs{
			JobId:       job.jobId,
			Device:      device,
			Target:      job.targetNode,
			Sync:        "full",
			AutoDismiss: false,
		},
	})
	if err != nil || cmdErr != nil {
		_, _ = conn.run(&Command{
			Execute: "blockdev-del",
			Arguments: &CommandNode{
				NodeName: job.targetNode,
			},
		})

		if err == nil {
			err = apiError(cmdErr)
		}
		job = nil
		return
	}

	return
}

func (j *BackupJob) find() (status *JobStatus, err error) {
	cmd := &Command{
		Execute: "query-jobs",
	}

	jobs := &JobStatusReturn{}
	err = RunCommand(j.vmId, cmd, jobs)
	if err != nil {
		return
	}

	if jobs.Error != nil {
		err = apiError(jobs.Error)
		return
	}

	for _, job := range jobs.Return {
		if job.Id == j.jobId {
			status = job
			return
		}
	}

	return
}

func (j *BackupJob) query() (status *JobStatus, err error) {
	status, err = j.find()
	if err != nil {
		return
	}

	if status == nil {
		err = &errortypes.ApiError{
			errors.Newf("qmp: Backup job %s missing", j.jobId),
		}
		return
	}

	return
}

func (j *BackupJob) release() (err error) {
	cmdErr, err := runChecked(j.vmId, &Command{
		Execute: "job-dismiss",
		Arguments: &CommandId{
			Id: j.jobId,
		},
	})
	if err != nil {
		return
	}

	if cmdErr != nil && !isNotFoundError(cmdErr) {
		err = apiError(cmdErr)
		return
	}

	cmdErr, err = runChecked(j.vmId, &Command{
		Execute: "blockdev-del",
		Arguments: &CommandNode{
			NodeName: j.targetNode,
		},
	})
	if err != nil {
		return
	}

	if cmdErr != nil && !isNotFoundError(cmdErr) {
		err = apiError(cmdErr)
		return
	}

	return
}

func (j *BackupJob) Wait() (err error) {
	var status *JobStatus
	lastLog := time.Now()

	for {
		status, err = j.query()
		if err != nil {
			return
		}

		if status.Status == "concluded" {
			break
		}

		if time.Since(lastLog) >= backupLogInterval {
			lastLog = time.Now()

			progress := 0
			if status.TotalProgress > 0 {
				progress = int(status.CurrentProgress * 100 /
					status.TotalProgress)
			}

			logrus.WithFields(logrus.Fields{
				"instance_id": j.vmId.Hex(),
				"disk_id":     j.dskId.Hex(),
				"status":      status.Status,
				"progress":    progress,
			}).Info("qmp: Disk backup in progress")
		}

		time.Sleep(backupPollInterval)
	}

	err = j.release()
	if err != nil {
		return
	}

	if status.Error != "" {
		err = &errortypes.WriteError{
			errors.Newf("qmp: Disk backup failed %s", status.Error),
		}
		return
	}

	logrus.WithFields(logrus.Fields{
		"instance_id": j.vmId.Hex(),
		"disk_id":     j.dskId.Hex(),
	}).Info("qmp: Disk backup complete")

	return
}

package qmp

import (
	"fmt"
	"strings"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/features"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/vm"
)

type blockDevFileArgs struct {
	Driver   string            `json:"driver"`
	NodeName string            `json:"node-name"`
	Aio      string            `json:"aio"`
	Discard  string            `json:"discard"`
	Filename string            `json:"filename"`
	Cache    blockDevFileCache `json:"cache"`
}

type blockDevFileCache struct {
	NoFlush bool `json:"no-flush"`
	Direct  bool `json:"direct"`
}

type blockDevArgs struct {
	Driver   string `json:"driver"`
	NodeName string `json:"node-name"`
	File     string `json:"file"`
}

type deviceAddArgs struct {
	Id     string `json:"id"`
	Driver string `json:"driver"`
	Drive  string `json:"drive"`
	Bus    string `json:"bus"`
}

type blockDevEventData struct {
	Device string `json:"device"`
	Path   string `json:"path"`
}

type blockDevEvent struct {
	Event string            `json:"event"`
	Data  blockDevEventData `json:"data"`
}

func AddDisk(vmId primitive.ObjectID, dsk *vm.Disk, virt *vm.VirtualMachine) (
	err error) {

	dskId := fmt.Sprintf("fd_%s", dsk.Id.Hex())
	dskFileId := fmt.Sprintf("fdf_%s", dsk.Id.Hex())
	dskDevId := fmt.Sprintf("fdd_%s", dsk.Id.Hex())

	diskAio := settings.Hypervisor.DiskAio
	if diskAio == "" {
		supported, e := features.GetUringSupport()
		if e != nil {
			err = e
			return
		}

		if supported {
			diskAio = "io_uring"
		} else {
			diskAio = "threads"
		}
	}

	conn := NewConnection(vmId)
	defer conn.Close()

	err = conn.Connect()
	if err != nil {
		return
	}

	cmd := &Command{
		Execute: "blockdev-add",
		Arguments: &blockDevFileArgs{
			Driver:   "file",
			NodeName: dskFileId,
			Aio:      diskAio,
			Discard:  "unmap",
			Filename: dsk.Path,
			Cache: blockDevFileCache{
				NoFlush: false,
				Direct:  true,
			},
		},
	}

	returnData := &CommandReturn{}
	err = conn.Send(cmd, returnData)
	if err != nil {
		return
	}

	if returnData.Error != nil &&
		!strings.Contains(
			strings.ToLower(returnData.Error.Desc),
			"duplicate",
		) {

		err = &errortypes.ApiError{
			errors.Newf("qmp: Return error %s", returnData.Error.Desc),
		}
		return
	}

	cmd = &Command{
		Execute: "blockdev-add",
		Arguments: &blockDevArgs{
			Driver:   "qcow2",
			NodeName: dskId,
			File:     dskFileId,
		},
	}

	returnData = &CommandReturn{}
	err = conn.Send(cmd, returnData)
	if err != nil {
		return
	}

	if returnData.Error != nil &&
		!strings.Contains(
			strings.ToLower(returnData.Error.Desc),
			"duplicate",
		) {

		err = &errortypes.ApiError{
			errors.Newf("qmp: Return error %s", returnData.Error.Desc),
		}
		return
	}

	cmd = &Command{
		Execute: "device_add",
		Arguments: &deviceAddArgs{
			Id:     dskDevId,
			Driver: "virtio-blk-pci",
			Drive:  dskId,
			Bus:    fmt.Sprintf("diskbus%d", dsk.Index),
		},
	}

	returnData = &CommandReturn{}
	err = conn.Send(cmd, returnData)
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

func RemoveDisk(vmId primitive.ObjectID, dsk *vm.Disk) (err error) {
	dskId := fmt.Sprintf("fd_%s", dsk.Id.Hex())
	dskFileId := fmt.Sprintf("fdf_%s", dsk.Id.Hex())
	dskDevId := fmt.Sprintf("fdd_%s", dsk.Id.Hex())

	diskAio := settings.Hypervisor.DiskAio
	if diskAio == "" {
		supported, e := features.GetUringSupport()
		if e != nil {
			err = e
			return
		}

		if supported {
			diskAio = "io_uring"
		} else {
			diskAio = "threads"
		}
	}

	conn := NewConnection(vmId)
	defer conn.Close()

	conn.SetDeadline(30 * time.Second)

	err = conn.Connect()
	if err != nil {
		return
	}

	cmd := &Command{
		Execute: "device_del",
		Arguments: &CommandId{
			Id: dskDevId,
		},
	}

	returnData := &CommandReturn{}
	err = conn.Send(cmd, returnData)
	if err != nil {
		return
	}

	skipEvent := false
	if returnData.Error != nil && (strings.Contains(
		strings.ToLower(returnData.Error.Desc),
		"process of unplug") || strings.Contains(
		strings.ToLower(returnData.Error.Desc),
		"not found") || strings.Contains(
		strings.ToLower(returnData.Error.Desc),
		"failed to find")) {

		skipEvent = true
	} else if returnData.Error != nil {
		err = &errortypes.ApiError{
			errors.Newf("qmp: Return error %s", returnData.Error.Desc),
		}
		return
	}

	if !skipEvent {
		event := &blockDevEvent{}
		err = conn.Event(event, func() (resp interface{}, err error) {
			if event.Event == "DEVICE_DELETED" &&
				event.Data.Device == dskDevId {

				return
			}

			event = &blockDevEvent{}
			resp = event

			return
		})
		if err != nil {
			return
		}
	}

	cmd = &Command{
		Execute: "blockdev-del",
		Arguments: &CommandNode{
			NodeName: dskId,
		},
	}

	returnData = &CommandReturn{}
	err = conn.Send(cmd, returnData)
	if err != nil {
		return
	}

	if returnData.Error != nil && !strings.Contains(
		strings.ToLower(returnData.Error.Desc),
		"process of unplug") && !strings.Contains(
		strings.ToLower(returnData.Error.Desc),
		"not found") && !strings.Contains(
		strings.ToLower(returnData.Error.Desc),
		"failed to find") {

		err = &errortypes.ApiError{
			errors.Newf("qmp: Return error %s", returnData.Error.Desc),
		}
		return
	}

	cmd = &Command{
		Execute: "blockdev-del",
		Arguments: &CommandNode{
			NodeName: dskFileId,
		},
	}

	returnData = &CommandReturn{}
	err = conn.Send(cmd, returnData)
	if err != nil {
		return
	}

	if returnData.Error != nil && !strings.Contains(
		strings.ToLower(returnData.Error.Desc),
		"process of unplug") && !strings.Contains(
		strings.ToLower(returnData.Error.Desc),
		"not found") && !strings.Contains(
		strings.ToLower(returnData.Error.Desc),
		"failed to find") {

		err = &errortypes.ApiError{
			errors.Newf("qmp: Return error %s", returnData.Error.Desc),
		}
		return
	}

	return
}

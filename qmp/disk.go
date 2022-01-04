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

type blockdevArgs struct {
	NodeName string        `json:"node-name"`
	Driver   string        `json:"driver"`
	Discard  string        `json:"discard"`
	Cache    blockDevCache `json:"cache"`
	File     blockDevFile  `json:"file"`
}

type blockDevFile struct {
	Driver   string `json:"driver"`
	Aio      string `json:"aio"`
	Filename string `json:"filename"`
}

type blockDevCache struct {
	NoFlush bool `json:"no-flush"`
	Direct  bool `json:"direct"`
}

type deviceAddArgs struct {
	Id     string `json:"id"`
	Driver string `json:"driver"`
	Drive  string `json:"drive"`
	Bus    string `json:"bus"`
}

func AddDisk(vmId primitive.ObjectID, dsk *vm.Disk, virt *vm.VirtualMachine) (
	err error) {

	dskId := fmt.Sprintf("disk_%s", dsk.Id.Hex())
	dskDevId := fmt.Sprintf("diskdev_%s", dsk.Id.Hex())

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
		Arguments: &blockdevArgs{
			NodeName: dskId,
			Driver:   "qcow2",
			Discard:  "unmap",
			Cache: blockDevCache{
				NoFlush: false,
				Direct:  true,
			},
			File: blockDevFile{
				Driver:   "file",
				Aio:      diskAio,
				Filename: dsk.Path,
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

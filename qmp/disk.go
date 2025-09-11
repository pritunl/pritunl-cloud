package qmp

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/features"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/sirupsen/logrus"
)

type blockDevFileArgs struct {
	Driver   string        `json:"driver"`
	NodeName string        `json:"node-name"`
	Aio      string        `json:"aio"`
	Discard  string        `json:"discard"`
	Filename string        `json:"filename"`
	Cache    blockDevCache `json:"cache"`
}

type blockDevArgs struct {
	Driver   string        `json:"driver"`
	NodeName string        `json:"node-name"`
	File     string        `json:"file"`
	Cache    blockDevCache `json:"cache"`
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

type blockDevEventData struct {
	Device string `json:"device"`
	Path   string `json:"path"`
}

type blockDevEvent struct {
	Event string            `json:"event"`
	Data  blockDevEventData `json:"data"`
}

func AddDisk(vmId bson.ObjectID, dsk *vm.Disk) (err error) {
	dskId := fmt.Sprintf("fd_%s", dsk.Id.Hex())
	dskFileId := fmt.Sprintf("fdf_%s", dsk.Id.Hex())
	dskDevId := fmt.Sprintf("fdd_%s", dsk.Id.Hex())

	logrus.WithFields(logrus.Fields{
		"instance_id": vmId.Hex(),
		"disk_id":     dsk.Id.Hex(),
		"disk_index":  dsk.Index,
	}).Info("qmp: Connecting virtual disk")

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

	conn := NewConnection(vmId, true)
	defer conn.Close()

	_, err = conn.Connect()
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
			Cache: blockDevCache{
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
			Cache: blockDevCache{
				NoFlush: false,
				Direct:  true,
			},
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

func RemoveDisk(vmId bson.ObjectID, dsk *vm.Disk) (err error) {
	dskId := fmt.Sprintf("fd_%s", dsk.Id.Hex())
	dskFileId := fmt.Sprintf("fdf_%s", dsk.Id.Hex())
	dskDevId := fmt.Sprintf("fdd_%s", dsk.Id.Hex())

	logrus.WithFields(logrus.Fields{
		"instance_id": vmId.Hex(),
		"disk_id":     dsk.Id.Hex(),
		"disk_index":  dsk.Index,
	}).Info("qmp: Disconnecting virtual disk")

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

	conn := NewConnection(vmId, true)
	defer conn.Close()

	conn.SetDeadline(30 * time.Second)

	_, err = conn.Connect()
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

type blockQueryReturn struct {
	Return []blockQueryDevice `json:"return"`
	Error  *CommandError      `json:"error"`
}

type blockQueryDevice struct {
	Device    string             `json:"device"`
	Locked    bool               `json:"locked"`
	Removable bool               `json:"removable"`
	Inserted  blockQueryInserted `json:"inserted"`
}

type blockQueryInserted struct {
	NodeName string          `json:"node-name"`
	Drv      string          `json:"drv"`
	File     string          `json:"file"`
	Cache    blockQueryCache `json:"cache"`
	Image    blockQueryImage `json:"image"`
}

type blockQueryCache struct {
	NoFlush   bool `json:"no-flush"`
	Direct    bool `json:"direct"`
	Writeback bool `json:"writeback"`
}

type blockQueryImage struct {
	VirtualSize int64  `json:"virtual-size"`
	Filename    string `json:"filename"`
	Format      string `json:"format"`
	ActualSize  int64  `json:"actual-size"`
}

type pciQueryReturn struct {
	Return []pciQueryBus `json:"return"`
	Error  *CommandError `json:"error"`
}

type pciQueryBus struct {
	Bus       int            `json:"bus"`
	Slot      int            `json:"slot"`
	QdevId    string         `json:"qdev_id"`
	Devices   []pciQueryBus  `json:"devices,omitempty"`
	PciBridge pciQueryBridge `json:"pci_bridge,omitempty"`
}

type pciQueryBridge struct {
	Devices []pciQueryDevice `json:"devices,omitempty"`
}

type pciQueryDevice struct {
	Bus    int    `json:"bus"`
	Slot   int    `json:"slot"`
	QdevId string `json:"qdev_id"`
}

func GetDisks(vmId bson.ObjectID) (info *QemuInfo, disks []*vm.Disk,
	err error) {

	conn := NewConnection(vmId, false)
	defer conn.Close()

	info, err = conn.Connect()
	if err != nil {
		return
	}

	cmd := &Command{
		Execute: "query-block",
	}

	returnData := &blockQueryReturn{}
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

	disksMap := map[bson.ObjectID]*vm.Disk{}

	index := 0
	for _, disk := range returnData.Return {
		var idSpl []string
		if strings.HasPrefix(disk.Device, "disk_") {
			idSpl = strings.Split(disk.Device, "_")
		} else if strings.HasPrefix(disk.Inserted.NodeName, "fd_") {
			idSpl = strings.Split(disk.Inserted.NodeName, "_")
		} else {
			continue
		}

		if len(idSpl) < 2 {
			logrus.WithFields(logrus.Fields{
				"instance_id":   vmId.Hex(),
				"qmp_names":     idSpl,
				"qmp_device":    disk.Device,
				"qmp_node_name": disk.Inserted.NodeName,
				"qmp_file":      disk.Inserted.File,
				"qmp_filename":  disk.Inserted.Image.Filename,
			}).Error("qmp: Disk id invalid")
			continue
		}

		dskId, ok := utils.ParseObjectId(idSpl[1])
		if !ok {
			logrus.WithFields(logrus.Fields{
				"instance_id":   vmId.Hex(),
				"qmp_names":     idSpl,
				"qmp_device":    disk.Device,
				"qmp_node_name": disk.Inserted.NodeName,
				"qmp_file":      disk.Inserted.File,
				"qmp_filename":  disk.Inserted.Image.Filename,
			}).Error("qmp: Disk id parse failed")
			continue
		}

		filename := disk.Inserted.Image.Filename
		if filename == "" {
			filename = disk.Inserted.File
			if filename == "" {
				logrus.WithFields(logrus.Fields{
					"instance_id":   vmId.Hex(),
					"qmp_names":     idSpl,
					"qmp_device":    disk.Device,
					"qmp_node_name": disk.Inserted.NodeName,
					"qmp_file":      disk.Inserted.File,
					"qmp_filename":  disk.Inserted.Image.Filename,
				}).Error("qmp: Disk filename invalid")
				continue
			}
		}

		dsk := &vm.Disk{
			Id:    dskId,
			Index: index,
			Path:  filename,
		}
		disks = append(disks, dsk)
		disksMap[dsk.Id] = dsk

		index += 1
	}

	cmd = &Command{
		Execute: "query-pci",
	}

	pciReturnData := &pciQueryReturn{}
	err = conn.Send(cmd, pciReturnData)
	if err != nil {
		return
	}

	if pciReturnData.Error != nil {
		err = &errortypes.ApiError{
			errors.Newf("qmp: Return error %s", pciReturnData.Error.Desc),
		}
		return
	}

	for _, rootBus := range pciReturnData.Return {
		if rootBus.Devices == nil {
			continue
		}
		for _, subBus := range rootBus.Devices {
			if !strings.HasPrefix(subBus.QdevId, "diskbus") ||
				subBus.PciBridge.Devices == nil {

				continue
			}
			for _, device := range subBus.PciBridge.Devices {
				if !strings.HasPrefix(device.QdevId, "fdd_") {
					continue
				}

				dskIndex, e := strconv.Atoi(subBus.QdevId[7:])
				if e != nil {
					logrus.WithFields(logrus.Fields{
						"instance_id": vmId.Hex(),
						"qmp_diskbus": subBus.QdevId,
						"qmp_device":  device.QdevId,
					}).Error("qmp: Disk bus parse failed")
					continue
				}

				dskId, ok := utils.ParseObjectId(device.QdevId[4:])
				if !ok {
					logrus.WithFields(logrus.Fields{
						"instance_id": vmId.Hex(),
						"qmp_diskbus": subBus.QdevId,
						"qmp_device":  device.QdevId,
					}).Error("qmp: Disk bus id parse failed")
					continue
				}

				dsk := disksMap[dskId]
				if dsk == nil {
					logrus.WithFields(logrus.Fields{
						"instance_id": vmId.Hex(),
						"disk_id":     dskId.Hex(),
						"qmp_diskbus": subBus.QdevId,
						"qmp_device":  device.QdevId,
					}).Error("qmp: Unknown disk found")
					continue
				}

				dsk.Index = dskIndex
			}
		}
	}

	return
}

package data

import (
	"encoding/json"
	"fmt"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/sirupsen/logrus"
)

type diskInfo struct {
	Filename    string `json:"filename"`
	Format      string `json:"format"`
	ActualSize  int    `json:"actual-size"`
	VirtualSize int    `json:"virtual-size"`
}

func GetDiskSize(dsk *disk.Disk) (size int, err error) {
	dskPth := paths.GetDiskPath(dsk.Id)

	output, err := utils.ExecOutput("",
		"qemu-img", "info", "--output=json", dskPth)
	if err != nil {
		return
	}

	diskInfo := &diskInfo{}

	err = json.Unmarshal([]byte(output), diskInfo)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "data: Failed to parse qemu disk info"),
		}
		return
	}

	size = diskInfo.VirtualSize / 1073741824

	return
}

func ExpandDisk(db *database.Database, dsk *disk.Disk) (err error) {
	dskPth := paths.GetDiskPath(dsk.Id)

	logrus.WithFields(logrus.Fields{
		"disk_id":   dsk.Id.Hex(),
		"disk_path": dskPth,
		"new_size":  dsk.NewSize,
	}).Info("data: Expanding disk")

	curSize, err := GetDiskSize(dsk)
	if err != nil {
		return
	}

	if curSize >= dsk.NewSize {
		logrus.WithFields(logrus.Fields{
			"disk_id":      dsk.Id.Hex(),
			"disk_path":    dskPth,
			"current_size": curSize,
			"new_size":     dsk.NewSize,
		}).Warn("data: Disk size larger then new size")

		dsk.Size = curSize
		return
	}

	expandSize := dsk.NewSize - curSize
	_, err = utils.ExecCombinedOutputLogged(nil,
		"qemu-img", "resize", dskPth, fmt.Sprintf("+%dG", expandSize))
	if err != nil {
		return
	}

	curSize, err = GetDiskSize(dsk)
	if err != nil {
		return
	}
	dsk.Size = curSize

	return
}

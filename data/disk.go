package data

import (
	"fmt"

	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/utils"
)

func CreateDisk(db *database.Database, dsk *disk.Disk) (
	backingImage string, err error) {

	diskPath := paths.GetDiskPath(dsk.Id)

	if !dsk.Image.IsZero() {
		backingImage, err = WriteImage(
			db, dsk.Image, dsk.Id, dsk.Size, dsk.Backing)
		if err != nil {
			return
		}
	} else {
		err = utils.Exec("", "qemu-img", "create",
			"-f", "qcow2", diskPath, fmt.Sprintf("%dG", dsk.Size))
		if err != nil {
			return
		}

		err = utils.Chmod(diskPath, 0600)
		if err != nil {
			return
		}
	}

	return
}

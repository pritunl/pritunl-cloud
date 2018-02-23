package data

import (
	"fmt"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
)

func CreateDisk(db *database.Database, dsk *disk.Disk) (err error) {
	diskPath := vm.GetDiskPath(dsk.Id)

	if dsk.Image != "" {
		err = WriteImage(db, dsk.Image, dsk.Id)
		if err != nil {
			return
		}
	} else {
		err = utils.Exec("", "qemu-img", "create",
			"-f", "qcow2", diskPath, fmt.Sprintf("%dG", dsk.Size))
		if err != nil {
			return
		}
	}

	return
}

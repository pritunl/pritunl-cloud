package data

import (
	"fmt"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/lvm"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/pool"
	"github.com/pritunl/pritunl-cloud/utils"
)

func createDiskQcow(db *database.Database, dsk *disk.Disk) (
	newSize int, backingImage string, err error) {

	diskPath := paths.GetDiskPath(dsk.Id)

	if !dsk.Image.IsZero() {
		newSize, backingImage, err = writeImageQcow(db, dsk)
		if err != nil {
			return
		}
	} else if dsk.FileSystem != "" {
		err = writeFsQcow(db, dsk)
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

func createDiskLvm(db *database.Database, dsk *disk.Disk) (
	newSize int, err error) {

	pl, err := pool.Get(db, dsk.Pool)
	if err != nil {
		return
	}

	err = lvm.InitLock(pl.VgName)
	if err != nil {
		return
	}

	if !dsk.Image.IsZero() {
		newSize, err = writeImageLvm(db, dsk, pl)
		if err != nil {
			return
		}
	} else if dsk.FileSystem != "" {
		err = writeFsLvm(db, dsk, pl)
		if err != nil {
			return
		}
	} else {
		err = lvm.CreateLv(pl.VgName, dsk.Id.Hex(), dsk.Size)
		if err != nil {
			return
		}
	}

	return
}

func CreateDisk(db *database.Database, dsk *disk.Disk) (
	newSize int, backingImage string, err error) {

	switch dsk.Type {
	case disk.Lvm:
		newSize, err = createDiskLvm(db, dsk)
		if err != nil {
			return
		}
		break
	case "", disk.Qcow2:
		newSize, backingImage, err = createDiskQcow(db, dsk)
		if err != nil {
			return
		}
		break
	default:
		err = &errortypes.ParseError{
			errors.Newf("data: Unknown disk type %s", dsk.Type),
		}
		return
	}

	return
}

func ActivateDisk(db *database.Database, dsk *disk.Disk) (err error) {
	if dsk.Type != disk.Lvm {
		return
	}

	pl, err := pool.Get(db, dsk.Pool)
	if err != nil {
		return
	}

	vgName := pl.VgName
	lvName := dsk.Id.Hex()

	err = lvm.ActivateLv(vgName, lvName)
	if err != nil {
		return
	}

	return
}

func DeactivateDisk(db *database.Database, dsk *disk.Disk) (err error) {
	if dsk.Type != disk.Lvm {
		return
	}

	pl, err := pool.Get(db, dsk.Pool)
	if err != nil {
		return
	}

	vgName := pl.VgName
	lvName := dsk.Id.Hex()

	err = lvm.DeactivateLv(vgName, lvName)
	if err != nil {
		return
	}

	return
}

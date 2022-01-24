package vmdk

import (
	"bytes"
	"os"

	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

func SetRandUuid(diskPath string) (err error) {
	diskUuid := uuid.New()

	diskFile, err := os.OpenFile(diskPath, os.O_RDWR, 0)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "vmdk: Failed to open file"),
		}
		return
	}
	defer func() {
		err = diskFile.Close()
		if err != nil {
			err = &errortypes.WriteError{
				errors.Wrap(err, "vmdk: Failed to write file"),
			}
			return
		}
	}()

	buffer := make([]byte, 10000)
	nRead, err := diskFile.Read(buffer)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "vmdk: Failed to read file"),
		}
		return
	}

	i := bytes.Index(buffer, []byte("ddb.uuid.image="))

	newBuffer := append(buffer[:i+16], []byte(diskUuid.String())...)
	newBuffer = append(newBuffer, buffer[i+52:]...)

	nWrite, err := diskFile.WriteAt(newBuffer, 0)
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "vmdk: Failed to write file"),
		}
		return
	}

	if nRead != nWrite {
		err = &errortypes.WriteError{
			errors.New("vmdk: Write count mismatch"),
		}
		return
	}

	err = diskFile.Sync()
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "vmdk: Failed to sync file"),
		}
		return
	}

	return
}

func SetUuid(diskPath string, diskUuid string) (err error) {
	diskFile, err := os.OpenFile(diskPath, os.O_RDWR, 0)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "vmdk: Failed to open file"),
		}
		return
	}
	defer func() {
		err = diskFile.Close()
		if err != nil {
			err = &errortypes.WriteError{
				errors.Wrap(err, "vmdk: Failed to write file"),
			}
			return
		}
	}()

	buffer := make([]byte, 10000)
	nRead, err := diskFile.Read(buffer)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "vmdk: Failed to read file"),
		}
		return
	}

	i := bytes.Index(buffer, []byte("ddb.uuid.image="))

	newBuffer := append(buffer[:i+16], []byte(diskUuid)...)
	newBuffer = append(newBuffer, buffer[i+52:]...)

	nWrite, err := diskFile.WriteAt(newBuffer, 0)
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "vmdk: Failed to write file"),
		}
		return
	}

	if nRead != nWrite {
		err = &errortypes.WriteError{
			errors.New("vmdk: Write count mismatch"),
		}
		return
	}

	err = diskFile.Sync()
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "vmdk: Failed to sync file"),
		}
		return
	}

	return
}

func GetUuid(diskPath string) (diskUuid string, err error) {
	diskFile, err := os.Open(diskPath)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "vmdk: Failed to open file"),
		}
		return
	}
	defer func() {
		err = diskFile.Close()
		if err != nil {
			err = &errortypes.WriteError{
				errors.Wrap(err, "vmdk: Failed to write file"),
			}
			return
		}
	}()

	buffer := make([]byte, 10000)
	_, err = diskFile.Read(buffer)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "vmdk: Failed to read file"),
		}
		return
	}

	i := bytes.Index(buffer, []byte("ddb.uuid.image="))

	diskUuid = string(buffer[i+16 : i+52])

	return
}

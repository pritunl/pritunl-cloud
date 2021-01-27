package drive

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

func GetDevices() (devices []*Device, err error) {
	if time.Since(syncLast) < 30*time.Second {
		devices = syncCache
		return
	}

	syncLock.Lock()
	defer syncLock.Unlock()

	diskIds, err := ioutil.ReadDir("/dev/disk/by-id/")
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "drive: Failed to list disk IDs"),
		}
		return
	}

	for _, item := range diskIds {
		filename := item.Name()

		device := &Device{
			Id: filename,
		}
		devices = append(devices, device)
	}

	syncCache = devices
	syncLast = time.Now()

	return
}

func GetDriveHashId(id string) string {
	hash := md5.New()
	hash.Write([]byte(id))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

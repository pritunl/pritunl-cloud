package qms

import (
	"fmt"
	"net"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

func VncPassword(vmId bson.ObjectID, passwd string) (err error) {
	sockPath, err := GetSockPath(vmId)
	if err != nil {
		return
	}

	lockId := socketsLock.Lock(vmId.Hex())
	defer socketsLock.Unlock(vmId.Hex(), lockId)

	conn, err := net.DialTimeout(
		"unix",
		sockPath,
		3*time.Second,
	)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "qemu: Failed to open socket"),
		}
		return
	}
	defer conn.Close()

	err = conn.SetDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "qemu: Failed set deadline"),
		}
		return
	}

	_, err = conn.Write(
		[]byte(fmt.Sprintf("change vnc password\n%s\n", passwd)))
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "qemu: Failed to write socket"),
		}
		return
	}

	time.Sleep(800 * time.Millisecond)

	return
}

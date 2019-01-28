package qms

import (
	"bytes"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"net"
	"strconv"
	"strings"
	"time"
)

var (
	socketsLock = utils.NewMultiTimeoutLock(1 * time.Minute)
)

func GetDisks(vmId primitive.ObjectID) (disks []*vm.Disk, err error) {
	sockPath := GetSockPath(vmId)
	disks = []*vm.Disk{}

	lockId := socketsLock.Lock(vmId.Hex())
	defer socketsLock.Unlock(vmId.Hex(), lockId)

	conn, err := net.DialTimeout(
		"unix",
		sockPath,
		1*time.Second,
	)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "qemu: Failed to open socket"),
		}
		return
	}
	defer conn.Close()

	err = conn.SetDeadline(time.Now().Add(2 * time.Second))
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "qemu: Failed set deadline"),
		}
		return
	}

	buffer := make([]byte, 100000)
	for {
		buf := make([]byte, 10000)
		n, e := conn.Read(buf)
		if e != nil {
			err = &errortypes.ReadError{
				errors.Wrap(e, "qemu: Failed to read socket"),
			}
			return
		}
		buffer = append(buffer, buf[:n]...)

		if bytes.Contains(bytes.TrimSpace(buffer), []byte("(qemu)")) {
			break
		}
	}

	_, err = conn.Write([]byte("info block\n"))
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "qemu: Failed to write socket"),
		}
		return
	}

	buffer = make([]byte, 100000)
	for {
		buf := make([]byte, 10000)
		n, e := conn.Read(buf)
		if e != nil {
			err = &errortypes.ReadError{
				errors.Wrap(e, "qemu: Failed to read socket"),
			}
			return
		}
		buffer = append(buffer, buf[:n]...)

		if bytes.Contains(bytes.TrimSpace(buffer), []byte("(qemu)")) {
			break
		}
	}

	for _, line := range strings.Split(string(buffer), "\n") {
		if !strings.HasPrefix(line, "virtio") || len(line) < 10 {
			continue
		}

		lineSpl := strings.SplitN(line[6:], ":", 2)
		if len(lineSpl) != 2 {
			logrus.WithFields(logrus.Fields{
				"instance_id": vmId.Hex(),
				"line":        line,
			}).Error("qemu: Unexpected qemu disk path")
			continue
		}

		index, e := strconv.Atoi(lineSpl[0])
		if e != nil {
			logrus.WithFields(logrus.Fields{
				"instance_id": vmId.Hex(),
				"line":        line,
			}).Error("qemu: Unexpected qemu disk path")
			continue
		}

		diskPath := strings.Fields(strings.TrimSpace(lineSpl[1]))[0]

		dsk := &vm.Disk{
			Index: index,
			Path:  diskPath,
		}
		disks = append(disks, dsk)
	}

	return
}

func AddDisk(vmId primitive.ObjectID, dsk *vm.Disk) (err error) {
	sockPath := GetSockPath(vmId)

	logrus.WithFields(logrus.Fields{
		"instance_id": vmId.Hex(),
		"disk_path":   dsk.Path,
	}).Info("qemu: Connecting virtual machine disk")

	lockId := socketsLock.Lock(vmId.Hex())
	defer socketsLock.Unlock(vmId.Hex(), lockId)

	conn, err := net.DialTimeout(
		"unix",
		sockPath,
		1*time.Second,
	)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "qemu: Failed to open socket"),
		}
		return
	}
	defer conn.Close()

	err = conn.SetDeadline(time.Now().Add(1 * time.Second))
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "qemu: Failed set deadline"),
		}
		return
	}

	drive := fmt.Sprintf(
		"file=%s,index=%d,media=disk,format=qcow2,discard=on,if=virtio\n",
		dsk.Path,
		dsk.Index,
	)

	_, err = conn.Write([]byte("drive_add virtio " + drive))
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "qemu: Failed to write socket"),
		}
		return
	}

	time.Sleep(1 * time.Second)

	return
}

func RemoveDisk(vmId primitive.ObjectID, dsk *vm.Disk) (err error) {
	sockPath := GetSockPath(vmId)

	logrus.WithFields(logrus.Fields{
		"instance_id": vmId.Hex(),
		"disk_path":   dsk.Path,
	}).Info("qemu: Disconnecting virtual machine disk")

	lockId := socketsLock.Lock(vmId.Hex())
	defer socketsLock.Unlock(vmId.Hex(), lockId)

	conn, err := net.DialTimeout(
		"unix",
		sockPath,
		1*time.Second,
	)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "qemu: Failed to open socket"),
		}
		return
	}
	defer conn.Close()

	err = conn.SetDeadline(time.Now().Add(1 * time.Second))
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "qemu: Failed set deadline"),
		}
		return
	}

	_, err = conn.Write([]byte(
		fmt.Sprintf("drive_del virtio%d\n", dsk.Index)))
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "qemu: Failed to write socket"),
		}
		return
	}

	time.Sleep(1 * time.Second)

	return
}

func Shutdown(vmId primitive.ObjectID) (err error) {
	sockPath := GetSockPath(vmId)

	lockId := socketsLock.Lock(vmId.Hex())
	defer socketsLock.Unlock(vmId.Hex(), lockId)

	conn, err := net.DialTimeout(
		"unix",
		sockPath,
		1*time.Second,
	)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "qemu: Failed to open socket"),
		}
		return
	}
	defer conn.Close()

	err = conn.SetDeadline(time.Now().Add(1 * time.Second))
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "qemu: Failed set deadline"),
		}
		return
	}

	_, err = conn.Write([]byte("system_powerdown\n"))
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "qemu: Failed to write socket"),
		}
		return
	}

	return
}

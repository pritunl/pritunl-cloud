package qms

import (
	"bytes"
	"fmt"
	"net"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/sirupsen/logrus"
)

func GetDisks(vmId primitive.ObjectID) (disks []*vm.Disk, err error) {
	disks = []*vm.Disk{}

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

	index := 0
	for _, line := range strings.Split(string(buffer), "\n") {
		if len(line) < 10 {
			continue
		}

		// TODO Backwards compatibility
		if strings.HasPrefix(line, "virtio") {
			line = strings.Replace(line, "\r", "", -1)

			if !strings.HasPrefix(line, "virtio") || len(line) < 10 {
				continue
			}
			line = strings.Replace(line, "\r", "", -1)

			lineSpl := strings.SplitN(line[6:], ":", 2)
			if len(lineSpl) != 2 {
				logrus.WithFields(logrus.Fields{
					"instance_id": vmId.Hex(),
					"line":        line,
				}).Error("qemu: Unexpected qemu disk path")
				continue
			}

			indexStr := strings.Fields(strings.TrimSpace(lineSpl[0]))[0]
			indx, e := strconv.Atoi(indexStr)
			if e != nil {
				logrus.WithFields(logrus.Fields{
					"instance_id": vmId.Hex(),
					"line":        line,
				}).Error("qemu: Unexpected qemu disk path index")
				continue
			}

			diskPath := strings.Fields(strings.TrimSpace(lineSpl[1]))[0]

			idStr := strings.Split(path.Base(diskPath), ".")[0]

			diskId, err := primitive.ObjectIDFromHex(idStr)
			if err != nil {
				continue
			}

			dsk := &vm.Disk{
				Id:    diskId,
				Index: indx,
				Path:  diskPath,
			}
			disks = append(disks, dsk)

			continue
		}

		if !strings.HasPrefix(line, "disk_") {
			continue
		}
		line = strings.Replace(line, "\r", "", -1)

		lineSpl := strings.SplitN(line, ":", 2)
		if len(lineSpl) != 2 {
			logrus.WithFields(logrus.Fields{
				"instance_id": vmId.Hex(),
				"line":        line,
			}).Error("qemu: Unexpected qemu disk id")
			continue
		}

		idIndexStr := strings.Fields(strings.TrimSpace(lineSpl[0]))[0]
		if len(idIndexStr) < 6 {
			logrus.WithFields(logrus.Fields{
				"instance_id": vmId.Hex(),
				"line":        line,
			}).Error("qemu: Unexpected qemu disk id length")
			continue
		}

		dskId, ok := utils.ParseObjectId(idIndexStr[5:])
		if !ok {
			logrus.WithFields(logrus.Fields{
				"instance_id": vmId.Hex(),
				"line":        line,
			}).Error("qemu: Unexpected qemu disk id parse")
			continue
		}

		diskPath := strings.Fields(strings.TrimSpace(lineSpl[1]))[0]

		dsk := &vm.Disk{
			Id:    dskId,
			Index: index,
			Path:  diskPath,
		}
		disks = append(disks, dsk)

		index += 1
	}

	return
}

func AddDisk(vmId primitive.ObjectID, dsk *vm.Disk) (err error) {
	sockPath, err := GetSockPath(vmId)
	if err != nil {
		return
	}

	logrus.WithFields(logrus.Fields{
		"instance_id": vmId.Hex(),
		"disk_path":   dsk.Path,
	}).Info("qemu: Connecting virtual machine disk")

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

	drive := fmt.Sprintf(
		"file=%s,index=%d,media=disk,format=qcow2,discard=off,if=virtio\n",
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
	sockPath, err := GetSockPath(vmId)
	if err != nil {
		return
	}

	logrus.WithFields(logrus.Fields{
		"instance_id": vmId.Hex(),
		"disk_path":   dsk.Path,
	}).Info("qemu: Disconnecting virtual machine disk")

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

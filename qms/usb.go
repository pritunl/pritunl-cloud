package qms

import (
	"bytes"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/usb"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/sirupsen/logrus"
)

func GetUsbDevices(vmId primitive.ObjectID) (
	devices []*vm.UsbDevice, err error) {

	sockPath := GetSockPath(vmId)

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

	_, err = conn.Write([]byte("info usb\n"))
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
		line = strings.TrimSpace(line)

		if !strings.HasPrefix(line, "Device") || len(line) < 10 {
			continue
		}
		line = strings.Replace(line, "\r", "", -1)

		if !strings.Contains(line, "ID:") {
			continue
		}

		lineSpl := strings.Split(line, "ID:")
		if len(lineSpl) != 2 {
			logrus.WithFields(logrus.Fields{
				"instance_id": vmId.Hex(),
				"line":        line,
			}).Error("qemu: Unexpected qemu usb info")
			continue
		}

		deviceId := strings.Fields(lineSpl[1])[0]

		device := &vm.UsbDevice{}
		if strings.HasPrefix(deviceId, "usbv") {
			lineSpl = strings.Split(deviceId[4:], "_")
			if len(lineSpl) != 2 {
				logrus.WithFields(logrus.Fields{
					"instance_id": vmId.Hex(),
					"line":        line,
				}).Error("qemu: Unexpected qemu usb id")
				continue
			}

			device.Vendor = lineSpl[0]
			device.Product = lineSpl[1]
		} else if strings.HasPrefix(deviceId, "usbb") {
			lineSpl = strings.Split(deviceId[4:], "_")
			if len(lineSpl) != 2 {
				logrus.WithFields(logrus.Fields{
					"instance_id": vmId.Hex(),
					"line":        line,
				}).Error("qemu: Unexpected qemu usb id")
				continue
			}

			device.Bus = lineSpl[0]
			device.Address = lineSpl[1]
		} else {
			logrus.WithFields(logrus.Fields{
				"instance_id": vmId.Hex(),
				"line":        line,
			}).Error("qemu: Unknown qemu usb id")
			continue
		}

		devices = append(devices, device)
	}

	return
}

func AddUsb(vmId primitive.ObjectID, device *vm.UsbDevice) (err error) {
	sockPath := GetSockPath(vmId)

	logrus.WithFields(logrus.Fields{
		"instance_id": vmId.Hex(),
		"usb_vendor":  device.Vendor,
		"usb_product": device.Product,
		"usb_bus":     device.Bus,
		"usb_address": device.Address,
	}).Info("qemu: Connecting virtual machine usb")

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

	vendor := usb.FilterId(device.Vendor)
	product := usb.FilterId(device.Product)
	bus := usb.FilterAddr(device.Bus)
	address := usb.FilterAddr(device.Address)

	deviceLine := ""
	if vendor != "" && product != "" {
		deviceLine = fmt.Sprintf(
			"usb-host,vendorid=0x%s,productid=0x%s,id=usbv%s_%s",
			vendor, product,
			vendor, product,
		)
	} else if bus != "" && address != "" {
		deviceLine = fmt.Sprintf(
			"usb-host,hostbus=%s,hostaddr=%s,id=usbb%s_%s",
			strings.TrimLeft(bus, "0"),
			strings.TrimLeft(address, "0"),
			bus,
			address,
		)
	} else {
		err = &errortypes.ReadError{
			errors.Wrap(err, "qemu: Unknown usb device id"),
		}
		return
	}

	_, err = conn.Write([]byte(
		fmt.Sprintf("device_add %s\n", deviceLine)))
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "qemu: Failed to write socket"),
		}
		return
	}

	time.Sleep(1 * time.Second)

	return
}

func RemoveUsb(vmId primitive.ObjectID, device *vm.UsbDevice) (err error) {
	sockPath := GetSockPath(vmId)

	logrus.WithFields(logrus.Fields{
		"instance_id": vmId.Hex(),
		"usb_vendor":  device.Vendor,
		"usb_product": device.Product,
		"usb_bus":     device.Bus,
		"usb_address": device.Address,
	}).Info("qemu: Disconnecting virtual machine usb")

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

	vendor := usb.FilterId(device.Vendor)
	product := usb.FilterId(device.Product)
	bus := usb.FilterAddr(device.Bus)
	address := usb.FilterAddr(device.Address)

	deviceId := ""
	if vendor != "" && product != "" {
		deviceId = fmt.Sprintf("usbv%s_%s", vendor, product)
	} else if bus != "" && address != "" {
		deviceId = fmt.Sprintf("usbb%s_%s", bus, address)
	} else {
		err = &errortypes.ReadError{
			errors.Wrap(err, "qemu: Unknown usb device id"),
		}
		return
	}

	_, err = conn.Write([]byte(
		fmt.Sprintf("device_del %s\n", deviceId)))
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "qemu: Failed to write socket"),
		}
		return
	}

	time.Sleep(1 * time.Second)

	return
}

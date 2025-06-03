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
	"github.com/pritunl/pritunl-cloud/permission"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/sirupsen/logrus"
)

func GetUsbDevices(vmId primitive.ObjectID) (
	devices []*vm.UsbDevice, err error) {

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

	err = conn.SetDeadline(time.Now().Add(3 * time.Second))
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "qemu: Failed set deadline"),
		}
		return
	}

	buffer := []byte{}
	for {
		buf := make([]byte, 5000000)
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

	buffer = []byte{}
	for {
		buf := make([]byte, 5000000)
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

		if strings.HasPrefix(deviceId, "usbd_") {
			lineSpl = strings.Split(deviceId, "_")
			if len(lineSpl) != 5 && len(lineSpl) != 6 {
				logrus.WithFields(logrus.Fields{
					"instance_id": vmId.Hex(),
					"line":        line,
				}).Error("qemu: Unexpected qemu usb id")
				continue
			}

			device := &vm.UsbDevice{
				Id:      deviceId,
				Bus:     lineSpl[1],
				Address: lineSpl[2],
				Vendor:  lineSpl[3],
				Product: lineSpl[4],
			}
			devices = append(devices, device)
		}
	}

	return
}

func AddUsb(virt *vm.VirtualMachine, device *vm.UsbDevice) (err error) {
	sockPath, err := GetSockPath(virt.Id)
	if err != nil {
		return
	}

	usbDevice, err := device.GetDevice()
	if err != nil {
		return
	}

	if usbDevice == nil {
		logrus.WithFields(logrus.Fields{
			"instance_id": virt.Id.Hex(),
			"usb_vendor":  device.Vendor,
			"usb_product": device.Product,
			"usb_bus":     device.Bus,
			"usb_address": device.Address,
		}).Warn("qemu: Failed to find usb device for attachment")
		return
	}

	if usbDevice.Bus == "" || usbDevice.Address == "" ||
		usbDevice.Vendor == "" || usbDevice.Product == "" {

		logrus.WithFields(logrus.Fields{
			"instance_id": virt.Id.Hex(),
			"usb_name":    usbDevice.Name,
			"usb_vendor":  usbDevice.Vendor,
			"usb_product": usbDevice.Product,
			"usb_bus":     usbDevice.Bus,
			"usb_address": usbDevice.Address,
			"usb_path":    usbDevice.BusPath,
		}).Warn("qemu: Failed to load usb device info for attachment")
		return
	}

	logrus.WithFields(logrus.Fields{
		"instance_id": virt.Id.Hex(),
		"usb_name":    usbDevice.Name,
		"usb_vendor":  usbDevice.Vendor,
		"usb_product": usbDevice.Product,
		"usb_bus":     usbDevice.Bus,
		"usb_address": usbDevice.Address,
		"usb_path":    usbDevice.BusPath,
	}).Info("qemu: Connecting virtual machine usb")

	err = permission.Chown(virt, usbDevice.BusPath)
	if err != nil {
		return
	}

	lockId := socketsLock.Lock(virt.Id.Hex())
	defer socketsLock.Unlock(virt.Id.Hex(), lockId)

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

	deviceLine := fmt.Sprintf(
		"usb-host,hostdevice=%s,id=%s",
		usbDevice.BusPath,
		usbDevice.GetQemuId(),
	)

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

func RemoveUsb(virt *vm.VirtualMachine, device *vm.UsbDevice) (err error) {
	sockPath, err := GetSockPath(virt.Id)
	if err != nil {
		return
	}

	usbDevice, err := device.GetDevice()
	if err != nil {
		return
	}

	if usbDevice != nil {
		logrus.WithFields(logrus.Fields{
			"instance_id": virt.Id.Hex(),
			"usb_id":      device.Id,
			"usb_name":    usbDevice.Name,
			"usb_vendor":  usbDevice.Vendor,
			"usb_product": usbDevice.Product,
			"usb_bus":     usbDevice.Bus,
			"usb_address": usbDevice.Address,
			"usb_path":    usbDevice.BusPath,
		}).Info("qemu: Disconnecting active usb device")
	} else {
		logrus.WithFields(logrus.Fields{
			"instance_id": virt.Id.Hex(),
			"usb_id":      device.Id,
			"usb_vendor":  device.Vendor,
			"usb_product": device.Product,
			"usb_bus":     device.Bus,
			"usb_address": device.Address,
		}).Info("qemu: Disconnecting inactive usb device")
	}

	lockId := socketsLock.Lock(virt.Id.Hex())
	defer socketsLock.Unlock(virt.Id.Hex(), lockId)

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

	if device.Id != "" {
		_, err = conn.Write([]byte(
			fmt.Sprintf("device_del %s\n", device.Id)))
		if err != nil {
			err = &errortypes.ReadError{
				errors.Wrap(err, "qemu: Failed to write socket"),
			}
			return
		}
	}

	time.Sleep(1 * time.Second)

	if usbDevice != nil && usbDevice.BusPath != "" {
		err = permission.Restore(usbDevice.BusPath)
		if err != nil {
			return
		}
	}

	return
}

package usb

import (
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/utils"
)

const (
	syncInterval = 6 * time.Second
)

var (
	syncLast               time.Time
	syncLock               sync.Mutex
	devicesCache           []*Device
	devicesIdMapCache      map[string]*Device
	devicesBusMapCache     map[string]*Device
	devicesBusPathMapCache map[string]*Device
)

type Device struct {
	Name       string `bson:"name" json:"name"`
	Vendor     string `bson:"vendor" json:"vendor"`
	Product    string `bson:"product" json:"product"`
	Bus        string `bson:"bus" json:"bus"`
	Address    string `bson:"address" json:"address"`
	DeviceName string `bson:"-" json:"-"`
	DevicePath string `bson:"-" json:"-"`
	BusPath    string `bson:"-" json:"-"`
}

func (d *Device) GetQemuId() string {
	return fmt.Sprintf("usbd_%s_%s_%s_%s_%d",
		d.Bus,
		d.Address,
		d.Vendor,
		d.Product,
		utils.RandInt(1111, 9999),
	)
}

func (d *Device) Unbind() (err error) {
	unbindPath := path.Join(d.DevicePath, "driver", "unbind")

	exists, err := utils.Exists(unbindPath)
	if err != nil {
		return
	}

	if exists {
		err = ioutil.WriteFile(
			unbindPath,
			[]byte(d.DeviceName),
			0644,
		)
		if err != nil {
			err = &errortypes.WriteError{
				errors.Wrapf(err, "usb: Failed to unbind '%s'", unbindPath),
			}
			return
		}
	}

	return
}

func syncDevices() (err error) {
	syncLock.Lock()
	defer syncLock.Unlock()

	devices := []*Device{}
	devicesIdMap := map[string]*Device{}
	devicesBusMap := map[string]*Device{}
	devicesBusPathMap := map[string]*Device{}
	basePath := "/sys/bus/usb/devices/"

	files, err := ioutil.ReadDir(basePath)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrapf(err, "usb: Failed to read dir '%s'", basePath),
		}
		return
	}

	for _, file := range files {
		devName := file.Name()
		if strings.Contains(devName, ":") ||
			strings.HasPrefix(devName, "usb") {

			continue
		}

		devPath := filepath.Join(basePath, devName)

		vendor, e := utils.ReadExists(filepath.Join(devPath, "idVendor"))
		if e != nil {
			err = e
			return
		}
		if vendor == "" {
			continue
		}

		product, e := utils.ReadExists(filepath.Join(devPath, "idProduct"))
		if e != nil {
			err = e
			return
		}
		if product == "" {
			continue
		}

		busNum, e := utils.ReadExists(filepath.Join(devPath, "busnum"))
		if e != nil {
			err = e
			return
		}
		if busNum == "" {
			continue
		}

		devNum, e := utils.ReadExists(filepath.Join(devPath, "devnum"))
		if e != nil {
			err = e
			return
		}
		if devNum == "" {
			continue
		}

		manufacturerDesc, e := utils.ReadExists(
			filepath.Join(devPath, "manufacturer"))
		if e != nil {
			err = e
			return
		}

		productDesc, e := utils.ReadExists(
			filepath.Join(devPath, "product"))
		if e != nil {
			err = e
			return
		}

		if manufacturerDesc == "" {
			manufacturerDesc = "Unknown Manufacturer"
		}
		if productDesc == "" {
			productDesc = "Unknown Product"
		}

		name := utils.FilterStr(strings.TrimSpace(manufacturerDesc)+
			" "+strings.TrimSpace(productDesc), 256)
		vendor = strings.TrimSpace(vendor)
		product = strings.TrimSpace(product)
		busNum = fmt.Sprintf("%03s", strings.TrimSpace(busNum))
		devNum = fmt.Sprintf("%03s", strings.TrimSpace(devNum))
		busPath := filepath.Join("/dev/bus/usb", busNum, devNum)

		device := &Device{
			Name:       name,
			Vendor:     vendor,
			Product:    product,
			Bus:        busNum,
			Address:    devNum,
			DeviceName: devName,
			DevicePath: devPath,
			BusPath:    busPath,
		}

		devices = append(devices, device)
		devicesIdMap[device.Vendor+":"+device.Product] = device
		devicesBusMap[device.Bus+"-"+device.Address] = device
		devicesBusPathMap[device.BusPath] = device
	}

	devicesCache = devices
	devicesIdMapCache = devicesIdMap
	devicesBusMapCache = devicesBusMap
	devicesBusPathMapCache = devicesBusPathMap
	syncLast = time.Now()

	return
}

func GetDevices() (devices []*Device, err error) {
	if time.Since(syncLast) > syncInterval {
		err = syncDevices()
		if err != nil {
			return
		}
	}

	syncLock.Lock()
	devices = devicesCache
	syncLock.Unlock()
	return
}

func GetDevice(bus, address, vendor, product string) (
	device *Device, err error) {

	if time.Since(syncLast) > syncInterval {
		err = syncDevices()
		if err != nil {
			return
		}
	}

	syncLock.Lock()
	if bus != "" && address != "" {
		device = devicesBusMapCache[bus+"-"+address]
		if device != nil && vendor != "" && product != "" {
			if device.Vendor != vendor || device.Product != product {
				device = nil
			}
		}
	} else {
		device = devicesIdMapCache[vendor+":"+product]
		if device != nil && bus != "" && address != "" {
			if device.Bus != bus || device.Address != address {
				device = nil
			}
		}
	}
	syncLock.Unlock()
	return
}

func GetDeviceId(vendor, product string) (device *Device, err error) {
	if time.Since(syncLast) > syncInterval {
		err = syncDevices()
		if err != nil {
			return
		}
	}

	syncLock.Lock()
	device = devicesIdMapCache[vendor+":"+product]
	syncLock.Unlock()
	return
}

func GetDeviceBus(bus, address string) (device *Device, err error) {
	if time.Since(syncLast) > syncInterval {
		err = syncDevices()
		if err != nil {
			return
		}
	}

	syncLock.Lock()
	device = devicesBusMapCache[bus+"-"+address]
	syncLock.Unlock()
	return
}

func GetDeviceBusPath(busPath string) (device *Device, err error) {
	if time.Since(syncLast) > syncInterval {
		err = syncDevices()
		if err != nil {
			return
		}
	}

	syncLock.Lock()
	device = devicesBusPathMapCache[busPath]
	syncLock.Unlock()
	return
}

package usb

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/utils"
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
	DevicePath string `bson:"-" json:"-"`
	BusPath    string `bson:"-" json:"-"`
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
		devPath := filepath.Join(basePath, devName)
		devNumPth := filepath.Join(devPath, "devnum")

		devNumExists, e := utils.Exists(devNumPth)
		if e != nil {
			err = e
			return
		}

		if !devNumExists {
			continue
		}

		vendor, e := utils.ReadExists(filepath.Join(devPath, "idVendor"))
		if e != nil {
			err = e
			return
		}

		product, e := utils.ReadExists(filepath.Join(devPath, "idProduct"))
		if e != nil {
			err = e
			return
		}

		busNum, e := utils.ReadExists(filepath.Join(devPath, "busnum"))
		if e != nil {
			err = e
			return
		}

		devNum, e := utils.ReadExists(devNumPth)
		if e != nil {
			err = e
			return
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

		if vendor == "" || product == "" || busNum == "" || devNum == "" {
			continue
		}

		if manufacturerDesc == "" {
			manufacturerDesc = "Unknown Manufacturer"
		}
		if productDesc == "" {
			productDesc = "Unknown Product"
		}

		name := utils.FilterStr(manufacturerDesc+" "+productDesc, 256)
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
			DevicePath: devPath,
			BusPath:    busPath,
		}

		devices = append(devices, device)
		devicesIdMap[device.Vendor+":"+device.Product] = device
		devicesIdMap[device.Bus+"-"+device.Address] = device
		devicesBusMap[device.BusPath] = device
	}

	devicesCache = devices
	devicesIdMapCache = devicesIdMap
	devicesBusMapCache = devicesBusMap
	devicesBusPathMapCache = devicesBusPathMap
	syncLast = time.Now()

	return
}

func GetDevices() (devices []*Device, err error) {
	if time.Since(syncLast) < 30*time.Second {
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

func GetDevice(vendor, product, bus, address string) (
	device *Device, err error) {

	if time.Since(syncLast) < 10*time.Second {
		err = syncDevices()
		if err != nil {
			return
		}
	}

	syncLock.Lock()
	if vendor != "" && product != "" {
		device = devicesIdMapCache[vendor+":"+product]
	} else {
		device = devicesBusMapCache[bus+"-"+address]
	}
	syncLock.Unlock()
	return
}

func GetDeviceId(vendor, product string) (device *Device, err error) {
	if time.Since(syncLast) < 10*time.Second {
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
	if time.Since(syncLast) < 10*time.Second {
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
	if time.Since(syncLast) < 10*time.Second {
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

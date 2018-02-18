package qemu

import (
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"net"
	"regexp"
	"strings"
	"sync"
	"time"
)

var (
	serviceReg  = regexp.MustCompile("pritunl_cloud_([a-z0-9]+).service")
	socketsLock = utils.NewMultiLock()
)

func GetVmInfo(vmId bson.ObjectId) (virt *vm.VirtualMachine, err error) {
	unitPath := GetUnitPath(vmId)

	data, err := ioutil.ReadFile(unitPath)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "qemu: Failed to read service"),
		}
		return
	}

	virt = &vm.VirtualMachine{}
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, "PritunlData=") {
			continue
		}

		lineSpl := strings.SplitN(line, "=", 2)
		if len(lineSpl) != 2 || len(lineSpl[1]) < 6 {
			continue
		}

		err = json.Unmarshal([]byte(lineSpl[1]), virt)
		if err != nil {
			err = &errortypes.ParseError{
				errors.Wrap(err, "qemu: Failed to parse service data"),
			}
			return
		}

		break
	}

	if virt.Id == "" {
		virt = nil
		return
	}

	unitName := GetUnitName(virt.Id)
	output, _ := utils.ExecOutput("", "systemctl", "is-active", unitName)
	state := strings.TrimSpace(output)

	switch state {
	case "active":
		virt.State = vm.Running
		break
	case "inactive":
		virt.State = vm.Stopped
		break
	case "failed":
		virt.State = vm.Failed
		break
	case "unknown":
		virt.State = vm.Stopped
		break
	default:
		logrus.WithFields(logrus.Fields{
			"id":    virt.Id.Hex(),
			"state": state,
		}).Info("qemu: Unknown virtual machine state")
		virt.State = vm.Failed
		break
	}

	return
}

func GetVms(db *database.Database) (virts []*vm.VirtualMachine, err error) {
	systemdPath := settings.Qemu.SystemdPath
	virts = []*vm.VirtualMachine{}

	items, err := ioutil.ReadDir(systemdPath)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "qemu: Failed to read systemd directory"),
		}
		return
	}

	units := []string{}
	for _, item := range items {
		if strings.HasPrefix(item.Name(), "pritunl_cloud") {
			units = append(units, item.Name())
		}
	}

	waiter := sync.WaitGroup{}
	virtsLock := sync.Mutex{}

	for _, unit := range units {
		match := serviceReg.FindStringSubmatch(unit)
		if match == nil || len(match) != 2 || !bson.IsObjectIdHex(match[1]) {
			continue
		}
		vmId := bson.ObjectIdHex(match[1])

		waiter.Add(1)
		go func() {
			defer waiter.Done()

			virt, e := GetVmInfo(vmId)
			if e != nil {
				err = e
				return
			}

			if virt != nil {
				e = virt.Commit(db)
				if e != nil {
					logrus.WithFields(logrus.Fields{
						"error": e,
					}).Error("qemu: Failed to commit VM state")
				}

				virtsLock.Lock()
				virts = append(virts, virt)
				virtsLock.Unlock()
			}
		}()
	}

	if err != nil {
		return
	}

	waiter.Wait()

	return
}

func Create(db *database.Database, virt *vm.VirtualMachine) (err error) {
	vmPath := vm.GetVmPath(virt.Id)
	unitName := GetUnitName(virt.Id)
	unitPath := GetUnitPath(virt.Id)

	logrus.WithFields(logrus.Fields{
		"id": virt.Id.Hex(),
	}).Info("qemu: Creating virtual machine")

	virt.State = "provisioning_disk"
	err = virt.Commit(db)
	if err != nil {
		return
	}

	err = utils.ExistsMkdir(vmPath, 0755)
	if err != nil {
		return
	}

	err = utils.Exec("", "cp", vm.ImagePath, virt.Disks[0].Path)
	if err != nil {
		return
	}

	qm, err := NewQemu(virt)
	if err != nil {
		return
	}

	output, err := qm.Marshal()
	if err != nil {
		return
	}

	err = utils.CreateWrite(unitPath, output, 0644)
	if err != nil {
		return
	}

	virt.State = vm.Starting
	err = virt.Commit(db)
	if err != nil {
		return
	}

	_, err = utils.ExecOutput("", "systemctl", "daemon-reload")
	if err != nil {
		return
	}

	output, err = utils.ExecOutput("", "systemctl", "start", unitName)
	if err != nil {
		return
	}

	return
}

func Update(db *database.Database, virt *vm.VirtualMachine) (err error) {
	unitPath := GetUnitPath(virt.Id)

	vrt, err := GetVmInfo(virt.Id)
	if err != nil {
		return
	}

	if vrt != nil && vrt.State != vm.Stopped && vrt.State != vm.Failed {
		err = &errortypes.WriteError{
			errors.Wrapf(err, "qemu: Cannot update running virtual machine"),
		}
		return
	}

	qm, err := NewQemu(virt)
	if err != nil {
		return
	}

	output, err := qm.Marshal()
	if err != nil {
		return
	}

	err = utils.CreateWrite(unitPath, output, 0644)
	if err != nil {
		return
	}

	_, err = utils.ExecOutput("", "systemctl", "daemon-reload")
	if err != nil {
		return
	}

	return
}

func Destroy(db *database.Database, virt *vm.VirtualMachine) (err error) {
	vmPath := vm.GetVmPath(virt.Id)
	unitName := GetUnitName(virt.Id)
	unitPath := GetUnitPath(virt.Id)
	sockPath := GetSockPath(virt.Id)
	pidPath := GetPidPath(virt.Id)

	logrus.WithFields(logrus.Fields{
		"id": virt.Id.Hex(),
	}).Info("qemu: Destroying virtual machine")

	_, err = utils.ExecOutput("", "systemctl", "stop", unitName)
	if err != nil {
		return
	}

	time.Sleep(3 * time.Second)

	err = utils.RemoveAll(vmPath)
	if err != nil {
		return
	}

	err = utils.RemoveAll(unitPath)
	if err != nil {
		return
	}

	err = utils.RemoveAll(sockPath)
	if err != nil {
		return
	}

	err = utils.RemoveAll(pidPath)
	if err != nil {
		return
	}

	err = utils.RemoveAll(unitPath)
	if err != nil {
		return
	}

	return
}

func PowerOn(db *database.Database, virt *vm.VirtualMachine) (err error) {
	unitName := GetUnitName(virt.Id)

	logrus.WithFields(logrus.Fields{
		"id": virt.Id.Hex(),
	}).Info("qemu: Power on virtual machine")

	_, err = utils.ExecOutput("", "systemctl", "start", unitName)
	if err != nil {
		return
	}

	time.Sleep(3 * time.Second)

	return
}

func sockPowerOff(virt *vm.VirtualMachine) (err error) {
	sockPath := GetSockPath(virt.Id)

	socketsLock.Lock(virt.Id.Hex())
	defer socketsLock.Unlock(virt.Id.Hex())

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

func PowerOff(db *database.Database, virt *vm.VirtualMachine) (err error) {
	unitName := GetUnitName(virt.Id)

	logrus.WithFields(logrus.Fields{
		"id": virt.Id.Hex(),
	}).Info("qemu: Power off virtual machine")

	for i := 0; i < 10; i++ {
		err = sockPowerOff(virt)
		if err == nil {
			break
		}

		time.Sleep(500 * time.Millisecond)
	}

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"id":    virt.Id.Hex(),
			"error": err,
		}).Error("qemu: Power off virtual machine error")
		err = nil
	} else {
		for i := 0; i < settings.Qemu.PowerOffTimeout; i++ {
			vrt, e := GetVmInfo(virt.Id)
			if e != nil {
				err = e
				return
			}

			if vrt == nil || vrt.State == vm.Stopped ||
				vrt.State == vm.Failed {

				if vrt != nil {
					err = vrt.Commit(db)
					if err != nil {
						return
					}
				}

				return
			}

			time.Sleep(1 * time.Second)
		}
	}

	logrus.WithFields(logrus.Fields{
		"id": virt.Id.Hex(),
	}).Warning("qemu: Force power off virtual machine")

	_, err = utils.ExecOutput("", "systemctl", "stop", unitName)
	if err != nil {
		return
	}

	time.Sleep(3 * time.Second)

	return
}

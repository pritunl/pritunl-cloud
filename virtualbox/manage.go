package virtualbox

import (
	"github.com/Sirupsen/logrus"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/pritunl/pritunl-cloud/vmdk"
	"gopkg.in/mgo.v2/bson"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	listReg = regexp.MustCompile("\"([a-z0-9]+)\" \\{(.*?)\\}")
	ipReg   = regexp.MustCompile(
		"Name: /VirtualBox/GuestInfo/Net/0/V4/IP, value: ([0-9]+\\.[0-9]+\\.[0-9]+\\.[0-9]+)")
	ip6Reg = regexp.MustCompile(
		"Name: /Pritunl/GuestInfo/Net/0/V6/IP, value: ([a-f0-9:]+:[a-f0-9:]+)")
)

func GetVmInfo(vmId bson.ObjectId) (virt *vm.VirtualMachine, err error) {
	virt = &vm.VirtualMachine{
		Id:              vmId,
		Disks:           []*vm.Disk{},
		NetworkAdapters: []*vm.NetworkAdapter{},
	}

	output, err := utils.ExecCombinedOutput("",
		ManageBin, "showvminfo", "--machinereadable", virt.Id.Hex())
	if err != nil {
		if strings.Contains(
			output, "Could not find a registered machine") {

			virt = nil
			err = nil
		}
		return
	}

	for _, line := range strings.Split(output, "\n") {
		lineSplt := strings.SplitN(line, "=", 2)
		if len(lineSplt) != 2 {
			continue
		}

		key := lineSplt[0]
		value := lineSplt[1]
		if strings.HasPrefix(value, "\"") &&
			strings.HasSuffix(value, "\"") {

			value = value[1 : len(value)-1]
		}

		if strings.HasPrefix(key, "bridgeadapter") {
			n, _ := strconv.Atoi(key[13:])
			if n < 1 {
				err = &errortypes.ParseError{
					errors.Newf(
						"virtualbox: Failed to get nic num %s",
						key,
					),
				}
				return
			}

			n -= 1

			if len(virt.NetworkAdapters) < n+1 {
				for i := 0; i < n+1-len(virt.NetworkAdapters); i++ {
					virt.NetworkAdapters = append(
						virt.NetworkAdapters,
						&vm.NetworkAdapter{},
					)
				}
			}

			virt.NetworkAdapters[n].BridgedInterface = value
		} else if strings.HasPrefix(key, "macaddress") {
			n, _ := strconv.Atoi(key[10:])
			if n < 1 {
				err = &errortypes.ParseError{
					errors.Newf(
						"virtualbox: Failed to get nic num %s",
						key,
					),
				}
				return
			}

			n -= 1

			if len(virt.NetworkAdapters) < n+1 {
				for i := 0; i < n+1-len(virt.NetworkAdapters); i++ {
					virt.NetworkAdapters = append(
						virt.NetworkAdapters,
						&vm.NetworkAdapter{},
					)
				}
			}

			virt.NetworkAdapters[n].MacAddress = value
		} else if strings.HasPrefix(key, "\"SATA-") &&
			strings.HasSuffix(key, "-0\"") {

			if value == "none" {
				continue
			}

			n, e := strconv.Atoi(key[6 : len(key)-3])
			if e != nil {
				continue
			}

			if n < 0 {
				err = &errortypes.ParseError{
					errors.Newf(
						"virtualbox: Failed to get disk num %s",
						key,
					),
				}
				return
			}

			if len(virt.Disks) < n+1 {
				for i := 0; i < n+1-len(virt.Disks); i++ {
					virt.Disks = append(
						virt.Disks,
						&vm.Disk{},
					)
				}
			}

			virt.Disks[n].Path = value
		} else {
			switch key {
			case "UUID":
				virt.Uuid = value
				break
			case "VMState":
				virt.State = value
				break
			case "cpus":
				cpus, _ := strconv.Atoi(value)
				virt.Processors = cpus
				break
			case "memory":
				memory, _ := strconv.Atoi(value)
				virt.Memory = memory
				break
			}
		}
	}

	output, err = utils.ExecOutput("",
		ManageBin, "guestproperty", "enumerate", virt.Id.Hex())
	if err != nil {
		return
	}

	match := ipReg.FindStringSubmatch(output)
	if match != nil || len(match) == 2 && len(virt.NetworkAdapters) > 0 {
		virt.NetworkAdapters[0].IpAddress = match[1]
	}

	match = ip6Reg.FindStringSubmatch(output)
	if match != nil || len(match) == 2 && len(virt.NetworkAdapters) > 0 {
		virt.NetworkAdapters[0].IpAddress6 = match[1]
	}

	return
}

func GetVms(db *database.Database) (virts []*vm.VirtualMachine, err error) {
	virts = []*vm.VirtualMachine{}

	output, err := utils.ExecOutput("", ManageBin, "list", "vms")
	if err != nil {
		return
	}

	waiter := sync.WaitGroup{}
	virtsLock := sync.Mutex{}

	for _, line := range strings.Split(output, "\n") {
		match := listReg.FindStringSubmatch(line)
		if match == nil || len(match) != 3 || !bson.IsObjectIdHex(match[1]) {
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
					}).Error("state: Failed to commit VM state")
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
	logrus.WithFields(logrus.Fields{
		"id": virt.Id.Hex(),
	}).Info("virtualbox: Creating virtual machine")

	vmPath := vm.GetVmPath(virt.Id)

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

	err = vmdk.SetRandUuid(virt.Disks[0].Path)
	if err != nil {
		return
	}

	vbox, err := NewVirtualBox(virt)
	if err != nil {
		return
	}

	output, err := vbox.Marshal()
	if err != nil {
		return
	}

	vboxPath := path.Join(vmPath, "vm.vbox")
	err = utils.CreateWrite(vboxPath, output, 0644)
	if err != nil {
		return
	}

	virt.State = "registering"
	err = virt.Commit(db)
	if err != nil {
		return
	}

	output, err = utils.ExecOutput("", ManageBin, "registervm", vboxPath)
	if err != nil {
		return
	}

	virt.State = "starting"
	err = virt.Commit(db)
	if err != nil {
		return
	}

	output, err = utils.ExecOutput("",
		ManageBin, "startvm", virt.Id.Hex(), "--type", "headless")
	if err != nil {
		return
	}

	return
}

func Update(db *database.Database, virt *vm.VirtualMachine) (err error) {
	curVirt, err := GetVmInfo(virt.Id)
	if err != nil {
		return
	}

	if curVirt == nil {
		err = &errortypes.NotFoundError{
			errors.Wrapf(err, "virtualbox: Failed to get VM info"),
		}
		return
	}

	if curVirt.State != vm.PowerOff {
		err = &errortypes.WriteError{
			errors.Wrapf(err, "virtualbox: Cannot update running VM"),
		}
		return
	}

	vmPath := vm.GetVmPath(virt.Id)

	virt.State = "updating"
	err = virt.Commit(db)
	if err != nil {
		return
	}

	_, err = utils.ExecOutput("",
		ManageBin, "unregistervm", virt.Id.Hex())
	if err != nil {
		return
	}

	vbox, err := NewVirtualBox(virt)
	if err != nil {
		return
	}

	output, err := vbox.Marshal()
	if err != nil {
		return
	}

	vboxPath := path.Join(vmPath, "vm.vbox")
	err = utils.CreateWrite(vboxPath, output, 0644)
	if err != nil {
		return
	}

	output, err = utils.ExecOutput("", ManageBin, "registervm", vboxPath)
	if err != nil {
		return
	}

	virt.State = vm.PowerOff
	err = virt.Commit(db)
	if err != nil {
		return
	}

	return
}

func Destroy(db *database.Database, virt *vm.VirtualMachine) (err error) {
	logrus.WithFields(logrus.Fields{
		"id": virt.Id.Hex(),
	}).Info("virtualbox: Destroying virtual machine")

	vmPath := vm.GetVmPath(virt.Id)

	_, err = utils.ExecOutput("",
		ManageBin, "controlvm", virt.Id.Hex(), "poweroff")
	if err != nil {
		return
	}

	time.Sleep(1500 * time.Millisecond)

	_, err = utils.ExecOutput("",
		ManageBin, "unregistervm", virt.Id.Hex(), "--delete")
	if err != nil {
		return
	}

	utils.RemoveAll(vmPath)

	return
}

func PowerOn(db *database.Database, virt *vm.VirtualMachine) (err error) {
	logrus.WithFields(logrus.Fields{
		"id": virt.Id.Hex(),
	}).Info("virtualbox: Power on virtual machine")

	_, err = utils.ExecOutput("",
		ManageBin, "startvm", virt.Id.Hex(), "--type", "headless")
	if err != nil {
		return
	}

	return
}

func PowerOff(virt *vm.VirtualMachine) (err error) {
	logrus.WithFields(logrus.Fields{
		"id": virt.Id.Hex(),
	}).Info("virtualbox: Power off virtual machine")

	_, err = utils.ExecOutput("",
		ManageBin, "controlvm", virt.Id.Hex(), "acpipowerbutton")
	if err != nil {
		return
	}

	return
}

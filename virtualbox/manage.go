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
	"time"
)

var (
	listReg = regexp.MustCompile("\"([a-z0-9]+)\" \\{(.*?)\\}")
	ipReg   = regexp.MustCompile(
		"Name: /VirtualBox/GuestInfo/Net/0/V4/IP, value: ([0-9]+\\.[0-9]+\\.[0-9]+\\.[0-9]+)")
	ip6Reg = regexp.MustCompile(
		"Name: /Pritunl/GuestInfo/Net/0/V6/IP, value: ([a-f0-9:]+:[a-f0-9:]+)")
)

func GetVms() (vms []*vm.VirtualMachine, err error) {
	vms = []*vm.VirtualMachine{}

	output, err := utils.ExecOutput("", ManageBin, "list", "vms")
	if err != nil {
		return
	}

	for _, line := range strings.Split(output, "\n") {
		match := listReg.FindStringSubmatch(line)
		if match == nil || len(match) != 3 || !bson.IsObjectIdHex(match[1]) {
			continue
		}

		v := &vm.VirtualMachine{
			Id:              bson.ObjectIdHex(match[1]),
			Uuid:            match[2],
			Disks:           []*vm.Disk{},
			NetworkAdapters: []*vm.NetworkAdapter{},
		}

		output, e := utils.ExecOutput("",
			ManageBin, "showvminfo", "--machinereadable", v.Id.Hex())
		if e != nil {
			err = e
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

				if len(v.NetworkAdapters) < n+1 {
					for i := 0; i < n+1-len(v.NetworkAdapters); i++ {
						v.NetworkAdapters = append(
							v.NetworkAdapters,
							&vm.NetworkAdapter{},
						)
					}
				}

				v.NetworkAdapters[n].BridgedInterface = value
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

				if len(v.NetworkAdapters) < n+1 {
					for i := 0; i < n+1-len(v.NetworkAdapters); i++ {
						v.NetworkAdapters = append(
							v.NetworkAdapters,
							&vm.NetworkAdapter{},
						)
					}
				}

				v.NetworkAdapters[n].MacAddress = value
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

				if len(v.Disks) < n+1 {
					for i := 0; i < n+1-len(v.Disks); i++ {
						v.Disks = append(
							v.Disks,
							&vm.Disk{},
						)
					}
				}

				v.Disks[n].Path = value
			} else {
				switch key {
				case "VMState":
					v.State = value
					break
				case "cpus":
					cpus, _ := strconv.Atoi(value)
					v.Processors = cpus
					break
				case "memory":
					memory, _ := strconv.Atoi(value)
					v.Memory = memory
					break
				}
			}
		}

		output, e = utils.ExecOutput("",
			ManageBin, "guestproperty", "enumerate", v.Id.Hex())
		if e != nil {
			err = e
			return
		}

		match = ipReg.FindStringSubmatch(output)
		if match != nil || len(match) == 2 && len(v.NetworkAdapters) > 0 {
			v.NetworkAdapters[0].IpAddress = match[1]
		}

		match = ip6Reg.FindStringSubmatch(output)
		if match != nil || len(match) == 2 && len(v.NetworkAdapters) > 0 {
			v.NetworkAdapters[0].IpAddress6 = match[1]
		}

		vms = append(vms, v)
	}

	return
}

func Create(virt *vm.VirtualMachine) (err error) {
	db := database.GetDatabase()
	defer db.Close()

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

	virt.State = "starting"
	err = virt.Commit(db)
	if err != nil {
		return
	}

	output, err = utils.ExecOutput("", ManageBin, "registervm", vboxPath)
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

func Destroy(virt *vm.VirtualMachine) (err error) {
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

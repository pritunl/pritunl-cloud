package qemu

import (
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/cloudinit"
	"github.com/pritunl/pritunl-cloud/data"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/iptables"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/qms"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/systemd"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/pritunl/pritunl-cloud/vpc"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"net"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	serviceReg = regexp.MustCompile("pritunl_cloud_([a-z0-9]+).service")
)

func GetVmInfo(vmId bson.ObjectId, getDisks bool) (
	virt *vm.VirtualMachine, err error) {

	unitPath := paths.GetUnitPath(vmId)

	unitData, err := ioutil.ReadFile(unitPath)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "qemu: Failed to read service"),
		}
		return
	}

	virt = &vm.VirtualMachine{}
	for _, line := range strings.Split(string(unitData), "\n") {
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

	unitName := paths.GetUnitName(virt.Id)
	state, err := systemd.GetState(unitName)
	if err != nil {
		return
	}

	switch state {
	case "active":
		virt.State = vm.Running
		break
	case "deactivating":
		virt.State = vm.Stopped
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

	if virt.State == vm.Running && getDisks {
		for i := 0; i < 10; i++ {
			disks, e := qms.GetDisks(vmId)
			if e != nil {
				if i < 9 {
					time.Sleep(250 * time.Millisecond)
					continue
				}
				err = e
				return
			}
			virt.Disks = disks

			break
		}
	}

	namespace := vm.GetNamespace(virt.Id, 0)
	ifaceInternal := vm.GetIfaceInternal(virt.Id, 0)

	ipData, err := utils.ExecCombinedOutputLogged(
		[]string{
			"No such file or directory",
			"does not exist",
			"setting the network namespace",
		},
		"ip", "netns", "exec", namespace,
		"ip", "-f", "inet", "-o", "addr",
		"show", "dev", ifaceInternal,
	)
	if err != nil {
		return
	}

	fields := strings.Fields(ipData)
	if len(fields) > 3 {
		ipAddr := net.ParseIP(strings.Split(fields[3], "/")[0])
		if ipAddr != nil && len(ipAddr) > 0 && len(virt.NetworkAdapters) > 0 {
			virt.NetworkAdapters[0].IpAddress = ipAddr.String()
		}
	}

	ipData, err = utils.ExecCombinedOutputLogged(
		[]string{
			"No such file or directory",
			"does not exist",
			"setting the network namespace",
		},
		"ip", "netns", "exec", namespace,
		"ip", "-f", "inet6", "-o", "addr",
		"show", "dev", ifaceInternal,
	)
	if err != nil {
		return
	}

	for _, line := range strings.Split(ipData, "\n") {
		if !strings.Contains(line, "global") {
			continue
		}

		fields = strings.Fields(ipData)
		if len(fields) > 3 {
			ipAddr := net.ParseIP(strings.Split(fields[3], "/")[0])
			if ipAddr != nil && len(ipAddr) > 0 &&
				len(virt.NetworkAdapters) > 0 {

				virt.NetworkAdapters[0].IpAddress6 = ipAddr.String()
			}
		}

		break
	}

	return
}

func GetVms(db *database.Database) (virts []*vm.VirtualMachine, err error) {
	systemdPath := settings.Hypervisor.SystemdPath
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

			virt, e := GetVmInfo(vmId, true)
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

func Wait(db *database.Database, virt *vm.VirtualMachine) (err error) {
	unitName := paths.GetUnitName(virt.Id)

	var vrt *vm.VirtualMachine
	for i := 0; i < settings.Hypervisor.StartTimeout; i++ {
		vrt, err = GetVmInfo(virt.Id, false)
		if err != nil {
			return
		}

		if vrt.State == vm.Running {
			break
		}
	}

	if vrt == nil || vrt.State != vm.Running {
		err = systemd.Stop(unitName)
		if err != nil {
			return
		}

		err = &errortypes.TimeoutError{
			errors.New("qemu: Power on timeout"),
		}

		return
	}

	return
}

func NetworkConf(db *database.Database, virt *vm.VirtualMachine) (err error) {
	ifaceNames := set.NewSet()

	if len(virt.NetworkAdapters) == 0 {
		err = &errortypes.NotFoundError{
			errors.New("qemu: Missing network interfaces"),
		}
		return
	}

	for i := range virt.NetworkAdapters {
		ifaceNames.Add(vm.GetIface(virt.Id, i))
	}

	for i := 0; i < 100; i++ {
		ifaces, e := net.Interfaces()
		if e != nil {
			err = &errortypes.ReadError{
				errors.Wrap(e, "qemu: Failed to get network interfaces"),
			}
			return
		}

		for _, iface := range ifaces {
			if ifaceNames.Contains(iface.Name) {
				ifaceNames.Remove(iface.Name)
			}
		}

		if ifaceNames.Len() == 0 {
			break
		}

		time.Sleep(250 * time.Millisecond)
	}

	if ifaceNames.Len() != 0 {
		err = &errortypes.ReadError{
			errors.New("qemu: Failed to find network interfaces"),
		}
		return
	}

	iface := vm.GetIface(virt.Id, 0)
	ifaceVirt := vm.GetIfaceVirt(virt.Id, 0)
	ifaceInternal := vm.GetIfaceInternal(virt.Id, 0)
	ifaceVlan := vm.GetIfaceVlan(virt.Id, 0)
	namespace := vm.GetNamespace(virt.Id, 0)
	pidPath := fmt.Sprintf("/var/run/dhclient-%s.pid", ifaceInternal)
	adapter := virt.NetworkAdapters[0]

	vc, err := vpc.Get(db, adapter.VpcId)
	if err != nil {
		return
	}

	vcNet, err := vc.GetNetwork()
	if err != nil {
		return
	}

	addr, err := vc.GetIp(db, vpc.Instance, virt.Id)
	if err != nil {
		return
	}

	gatewayAddr, err := vc.GetIp(db, vpc.Gateway, virt.Id)
	if err != nil {
		return
	}

	addr6 := vc.GetIp6(addr)
	if err != nil {
		return
	}

	gatewayAddr6 := vc.GetIp6(gatewayAddr)
	if err != nil {
		return
	}

	cidr, _ := vcNet.Mask.Size()
	gatewayCidr := fmt.Sprintf("%s/%d", gatewayAddr.String(), cidr)

	_, err = utils.ExecCombinedOutputLogged(
		[]string{"File exists"},
		"ip", "netns",
		"add", namespace,
	)
	if err != nil {
		PowerOff(db, virt)
		return
	}

	utils.ExecCombinedOutput("", "ip", "link", "set", ifaceVirt, "down")
	utils.ExecCombinedOutput("", "ip", "link", "del", ifaceVirt)

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "link",
		"add", ifaceVirt,
		"type", "veth",
		"peer", "name", ifaceInternal,
	)
	if err != nil {
		PowerOff(db, virt)
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "link",
		"set", "dev", ifaceVirt, "up",
	)
	if err != nil {
		PowerOff(db, virt)
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		[]string{"already a member of a bridge"},
		"brctl", "addif", adapter.HostInterface, ifaceVirt)
	if err != nil {
		PowerOff(db, virt)
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		[]string{"File exists"},
		"ip", "link",
		"set", "dev", ifaceInternal,
		"netns", namespace,
	)
	if err != nil {
		PowerOff(db, virt)
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", namespace,
		"sysctl", "-w", "net.ipv6.conf.all.accept_ra=2",
	)
	if err != nil {
		PowerOff(db, virt)
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", namespace,
		"sysctl", "-w", "net.ipv6.conf.default.accept_ra=2",
	)
	if err != nil {
		PowerOff(db, virt)
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", namespace,
		"sysctl", "-w",
		fmt.Sprintf("net.ipv6.conf.%s.accept_ra=2", ifaceInternal),
	)
	if err != nil {
		PowerOff(db, virt)
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", namespace,
		"sysctl", "-w", "net.ipv4.ip_forward=1",
	)
	if err != nil {
		PowerOff(db, virt)
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", namespace,
		"sysctl", "-w", "net.ipv6.conf.all.forwarding=1",
	)
	if err != nil {
		PowerOff(db, virt)
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		[]string{"File exists"},
		"ip", "link",
		"set", "dev", iface,
		"netns", namespace,
	)
	if err != nil {
		PowerOff(db, virt)
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", namespace,
		"ip", "link",
		"set", "dev", "lo", "up",
	)
	if err != nil {
		PowerOff(db, virt)
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", namespace,
		"ip", "link",
		"set", "dev", ifaceInternal, "up",
	)
	if err != nil {
		PowerOff(db, virt)
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", namespace,
		"ip", "link",
		"set", "dev", iface, "up",
	)
	if err != nil {
		PowerOff(db, virt)
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		[]string{"File exists"},
		"ip", "netns", "exec", namespace,
		"ip", "link",
		"add", "link", ifaceInternal,
		"name", ifaceVlan,
		"type", "vlan",
		"id", strconv.Itoa(vc.VpcId),
	)
	if err != nil {
		PowerOff(db, virt)
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", namespace,
		"ip", "link",
		"set", "dev", ifaceVlan, "up",
	)
	if err != nil {
		PowerOff(db, virt)
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		[]string{"already exists"},
		"ip", "netns", "exec", namespace,
		"brctl", "addbr", "br0",
	)
	if err != nil {
		PowerOff(db, virt)
		return
	}

	//_, err = utils.ExecCombinedOutputLogged(
	//	nil,
	//	"ip", "netns", "exec", namespace,
	//	"brctl", "stp", "br0", "on",
	//)
	//if err != nil {
	//	PowerOff(db, virt)
	//	return
	//}

	_, err = utils.ExecCombinedOutputLogged(
		[]string{"already a member of a bridge"},
		"ip", "netns", "exec", namespace,
		"brctl", "addif", "br0", ifaceVlan,
	)
	if err != nil {
		PowerOff(db, virt)
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		[]string{"already a member of a bridge"},
		"ip", "netns", "exec", namespace,
		"brctl", "addif", "br0", iface,
	)
	if err != nil {
		PowerOff(db, virt)
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		[]string{"File exists"},
		"ip", "netns", "exec", namespace,
		"ip", "addr",
		"add", gatewayCidr,
		"dev", "br0",
	)
	if err != nil {
		PowerOff(db, virt)
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		[]string{"File exists"},
		"ip", "netns", "exec", namespace,
		"ip", "-6", "addr",
		"add", gatewayAddr6.String()+"/64",
		"dev", "br0",
	)
	if err != nil {
		PowerOff(db, virt)
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", namespace,
		"ip", "link",
		"set", "dev", "br0", "up",
	)
	if err != nil {
		PowerOff(db, virt)
		return
	}

	networkStopDhClient(db, virt)

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", namespace,
		"dhclient", "-pf", pidPath,
		ifaceInternal,
	)
	if err != nil {
		PowerOff(db, virt)
		return
	}

	// TODO
	time.Sleep(2 * time.Second)

	ipData, err := utils.ExecCombinedOutputLogged(
		[]string{
			"No such file or directory",
			"does not exist",
		},
		"ip", "netns", "exec", namespace,
		"ip", "-f", "inet", "-o", "addr",
		"show", "dev", ifaceInternal,
	)
	if err != nil {
		return
	}

	pubAddr := ""
	fields := strings.Fields(ipData)
	if len(fields) > 3 {
		ipAddr := net.ParseIP(strings.Split(fields[3], "/")[0])
		if ipAddr != nil && len(ipAddr) > 0 && len(virt.NetworkAdapters) > 0 {
			pubAddr = ipAddr.String()
		}
	}

	ipData, err = utils.ExecCombinedOutputLogged(
		[]string{
			"No such file or directory",
			"does not exist",
		},
		"ip", "netns", "exec", namespace,
		"ip", "-f", "inet6", "-o", "addr",
		"show", "dev", ifaceInternal,
	)
	if err != nil {
		return
	}

	pubAddr6 := ""
	for _, line := range strings.Split(ipData, "\n") {
		if !strings.Contains(line, "global") {
			continue
		}

		fields = strings.Fields(ipData)
		if len(fields) > 3 {
			ipAddr := net.ParseIP(strings.Split(fields[3], "/")[0])
			if ipAddr != nil && len(ipAddr) > 0 &&
				len(virt.NetworkAdapters) > 0 {

				pubAddr6 = ipAddr.String()
			}
		}

		break
	}

	iptables.Lock()
	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", namespace,
		"iptables", "-t", "nat",
		"-A", "POSTROUTING",
		"-o", ifaceInternal,
		"-j", "MASQUERADE",
	)
	iptables.Unlock()
	if err != nil {
		PowerOff(db, virt)
		return
	}

	iptables.Lock()
	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", namespace,
		"iptables", "-t", "nat",
		"-A", "PREROUTING",
		"-d", pubAddr,
		"-j", "DNAT",
		"--to-destination", addr.String(),
	)
	iptables.Unlock()
	if err != nil {
		PowerOff(db, virt)
		return
	}

	if pubAddr6 != "" {
		iptables.Lock()
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "netns", "exec", namespace,
			"ip6tables", "-t", "nat",
			"-A", "POSTROUTING",
			"-o", ifaceInternal,
			"-j", "MASQUERADE",
		)
		iptables.Unlock()
		if err != nil {
			PowerOff(db, virt)
			return
		}

		iptables.Lock()
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "netns", "exec", namespace,
			"ip6tables", "-t", "nat",
			"-A", "PREROUTING",
			"-d", pubAddr6,
			"-j", "DNAT",
			"--to-destination", addr6.String(),
		)
		iptables.Unlock()
		if err != nil {
			PowerOff(db, virt)
			return
		}
	} else {
		logrus.WithFields(logrus.Fields{
			"instance_id":   virt.Id.Hex(),
			"net_namespace": namespace,
		}).Warning("qemu: Instance missing IPv6 address")
	}

	coll := db.Instances()
	err = coll.UpdateId(virt.Id, &bson.M{
		"$set": &bson.M{
			"private_ips":  []string{addr.String()},
			"private_ips6": []string{addr6.String()},
		},
	})
	if err != nil {
		err = database.ParseError(err)
		if _, ok := err.(*database.NotFoundError); ok {
			err = nil
		} else {
			return
		}
	}

	//_, err = utils.ExecCombinedOutputLogged(
	//	nil,
	//	"ip", "netns", "exec", namespace,
	//	"iptables", "-t", "nat",
	//	"-A", "POSTROUTING",
	//	"-s", vcAddr.String(),
	//	"-j", "SNAT",
	//	"--to-source", addr,
	//)
	//if err != nil {
	//	PowerOff(db, virt)
	//	return
	//}

	return
}

func networkStopDhClient(db *database.Database,
	virt *vm.VirtualMachine) (err error) {

	if len(virt.NetworkAdapters) == 0 {
		err = &errortypes.NotFoundError{
			errors.New("qemu: Missing network interfaces"),
		}
		return
	}

	ifaceInternal := vm.GetIfaceInternal(virt.Id, 0)
	pidPath := fmt.Sprintf("/var/run/dhclient-%s.pid", ifaceInternal)

	pid := ""
	pidData, _ := ioutil.ReadFile(pidPath)
	if pidData != nil {
		pid = strings.TrimSpace(string(pidData))
	}

	if pid != "" {
		utils.ExecCombinedOutput("", "kill", pid)
	}

	utils.RemoveAll(pidPath)

	return
}

func NetworkConfClear(db *database.Database,
	virt *vm.VirtualMachine) (err error) {

	if len(virt.NetworkAdapters) == 0 {
		err = &errortypes.NotFoundError{
			errors.New("qemu: Missing network interfaces"),
		}
		return
	}

	err = networkStopDhClient(db, virt)
	if err != nil {
		return
	}

	ifaceVirt := vm.GetIfaceVirt(virt.Id, 0)
	namespace := vm.GetNamespace(virt.Id, 0)

	utils.ExecCombinedOutput("", "ip", "netns", "del", namespace)
	utils.ExecCombinedOutput("", "ip", "link", "set", ifaceVirt, "down")
	utils.ExecCombinedOutput("", "ip", "link", "del", ifaceVirt)

	return
}

func writeService(virt *vm.VirtualMachine) (err error) {
	unitPath := paths.GetUnitPath(virt.Id)

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

	err = systemd.Reload()
	if err != nil {
		return
	}

	return
}

func Create(db *database.Database, inst *instance.Instance,
	virt *vm.VirtualMachine) (err error) {

	vmPath := paths.GetVmPath(virt.Id)
	unitName := paths.GetUnitName(virt.Id)

	logrus.WithFields(logrus.Fields{
		"id": virt.Id.Hex(),
	}).Info("qemu: Creating virtual machine")

	virt.State = vm.Provisioning
	err = virt.Commit(db)
	if err != nil {
		return
	}

	err = utils.ExistsMkdir(settings.Hypervisor.LibPath, 0755)
	if err != nil {
		return
	}

	err = utils.ExistsMkdir(vmPath, 0755)
	if err != nil {
		return
	}

	dsk, err := disk.GetInstanceIndex(db, inst.Id, "0")
	if err != nil {
		if _, ok := err.(*database.NotFoundError); ok {
			dsk = nil
			err = nil
		} else {
			return
		}
	}

	if dsk == nil {
		dsk = &disk.Disk{
			Id:             bson.NewObjectId(),
			Name:           inst.Name,
			State:          disk.Available,
			Node:           node.Self.Id,
			Organization:   inst.Organization,
			Instance:       inst.Id,
			SourceInstance: inst.Id,
			Image:          virt.Image,
			Index:          "0",
			Size:           inst.InitDiskSize,
		}

		err = data.WriteImage(db, virt.Image, dsk.Id, inst.InitDiskSize)
		if err != nil {
			return
		}

		err = dsk.Insert(db)
		if err != nil {
			return
		}
	}

	virt.Disks = append(virt.Disks, &vm.Disk{
		Index: 0,
		Path:  paths.GetDiskPath(dsk.Id),
	})

	err = cloudinit.Write(db, inst, virt)
	if err != nil {
		return
	}

	err = writeService(virt)
	if err != nil {
		return
	}

	virt.State = vm.Starting
	err = virt.Commit(db)
	if err != nil {
		return
	}

	err = systemd.Start(unitName)
	if err != nil {
		return
	}

	err = Wait(db, virt)
	if err != nil {
		return
	}

	err = NetworkConf(db, virt)
	if err != nil {
		return
	}

	return
}

func Destroy(db *database.Database, virt *vm.VirtualMachine) (err error) {
	vmPath := paths.GetVmPath(virt.Id)
	unitName := paths.GetUnitName(virt.Id)
	unitPath := paths.GetUnitPath(virt.Id)
	sockPath := paths.GetSockPath(virt.Id)
	guestPath := paths.GetGuestPath(virt.Id)
	pidPath := paths.GetPidPath(virt.Id)

	logrus.WithFields(logrus.Fields{
		"id": virt.Id.Hex(),
	}).Info("qemu: Destroying virtual machine")

	exists, err := utils.Exists(unitPath)
	if err != nil {
		return
	}

	if exists {
		err = systemd.Stop(unitName)
		if err != nil {
			return
		}
	}

	time.Sleep(3 * time.Second)

	err = NetworkConfClear(db, virt)
	if err != nil {
		return
	}

	for i, dsk := range virt.Disks {
		ds, e := disk.Get(db, dsk.GetId())
		if e != nil {
			err = e
			return
		}

		if i == 0 && ds.SourceInstance == virt.Id {
			err = disk.Delete(db, ds.Id)
			if err != nil {
				return
			}
		} else {
			err = disk.Detach(db, dsk.GetId())
			if err != nil {
				return
			}
		}
	}

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

	err = utils.RemoveAll(guestPath)
	if err != nil {
		return
	}

	err = utils.RemoveAll(pidPath)
	if err != nil {
		return
	}

	err = utils.RemoveAll(paths.GetInitPath(virt.Id))
	if err != nil {
		return
	}

	err = utils.RemoveAll(unitPath)
	if err != nil {
		return
	}

	return
}

func PowerOn(db *database.Database, inst *instance.Instance,
	virt *vm.VirtualMachine) (err error) {
	unitName := paths.GetUnitName(virt.Id)

	logrus.WithFields(logrus.Fields{
		"id": virt.Id.Hex(),
	}).Info("qemu: Starting virtual machine")

	err = cloudinit.Write(db, inst, virt)
	if err != nil {
		return
	}

	err = writeService(virt)
	if err != nil {
		return
	}

	err = systemd.Start(unitName)
	if err != nil {
		return
	}

	err = Wait(db, virt)
	if err != nil {
		return
	}

	err = NetworkConf(db, virt)
	if err != nil {
		return
	}

	return
}

func PowerOff(db *database.Database, virt *vm.VirtualMachine) (err error) {
	unitName := paths.GetUnitName(virt.Id)

	logrus.WithFields(logrus.Fields{
		"id": virt.Id.Hex(),
	}).Info("qemu: Stopping virtual machine")

	for i := 0; i < 10; i++ {
		err = qms.Shutdown(virt.Id)
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
		for i := 0; i < settings.Hypervisor.StopTimeout; i++ {
			vrt, e := GetVmInfo(virt.Id, false)
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

			if i != 0 && i%3 == 0 {
				qms.Shutdown(virt.Id)
			}

			time.Sleep(1 * time.Second)
		}
	}

	logrus.WithFields(logrus.Fields{
		"id": virt.Id.Hex(),
	}).Warning("qemu: Force power off virtual machine")

	err = systemd.Stop(unitName)
	if err != nil {
		return
	}

	err = NetworkConfClear(db, virt)
	if err != nil {
		return
	}

	time.Sleep(3 * time.Second)

	return
}

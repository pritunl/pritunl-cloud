package instance

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/gorilla/websocket"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/block"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/drive"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/iscsi"
	"github.com/pritunl/pritunl-cloud/iso"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/pci"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/systemd"
	"github.com/pritunl/pritunl-cloud/usb"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/pritunl/pritunl-cloud/vpc"
	"github.com/sirupsen/logrus"
)

type Instance struct {
	Id                  primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Organization        primitive.ObjectID `bson:"organization" json:"organization"`
	Zone                primitive.ObjectID `bson:"zone" json:"zone"`
	Vpc                 primitive.ObjectID `bson:"vpc" json:"vpc"`
	Subnet              primitive.ObjectID `bson:"subnet" json:"subnet"`
	OracleSubnet        string             `bson:"oracle_subnet" json:"oracle_subnet"`
	OracleVnic          string             `bson:"oracle_vnic" json:"oracle_vnic"`
	OracleVnicAttach    string             `bson:"oracle_vnic_attach" json:"oracle_vnic_attach"`
	Image               primitive.ObjectID `bson:"image" json:"image"`
	ImageBacking        bool               `bson:"image_backing" json:"image_backing"`
	Status              string             `bson:"-" json:"status"`
	Uptime              string             `bson:"-" json:"uptime"`
	State               string             `bson:"state" json:"state"`
	PublicMac           string             `bson:"-" json:"public_mac"`
	VmState             string             `bson:"vm_state" json:"vm_state"`
	VmTimestamp         time.Time          `bson:"vm_timestamp" json:"vm_timestamp"`
	Restart             bool               `bson:"restart" json:"restart"`
	RestartBlockIp      bool               `bson:"restart_block_ip" json:"restart_block_ip"`
	Uefi                bool               `bson:"uefi" json:"uefi"`
	SecureBoot          bool               `bson:"secure_boot" json:"secure_boot"`
	DeleteProtection    bool               `bson:"delete_protection" json:"delete_protection"`
	PublicIps           []string           `bson:"public_ips" json:"public_ips"`
	PublicIps6          []string           `bson:"public_ips6" json:"public_ips6"`
	PrivateIps          []string           `bson:"private_ips" json:"private_ips"`
	PrivateIps6         []string           `bson:"private_ips6" json:"private_ips6"`
	GatewayIps          []string           `bson:"gateway_ips" json:"gateway_ips"`
	GatewayIps6         []string           `bson:"gateway_ips6" json:"gateway_ips6"`
	OraclePrivateIps    []string           `bson:"oracle_private_ips" json:"oracle_private_ips"`
	OraclePublicIps     []string           `bson:"oracle_public_ips" json:"oracle_public_ips"`
	HostIps             []string           `bson:"host_ips" json:"host_ips"`
	NetworkNamespace    string             `bson:"network_namespace" json:"network_namespace"`
	NoPublicAddress     bool               `bson:"no_public_address" json:"no_public_address"`
	NoHostAddress       bool               `bson:"no_host_address" json:"no_host_address"`
	Node                primitive.ObjectID `bson:"node" json:"node"`
	Domain              primitive.ObjectID `bson:"domain,omitempty" json:"domain"`
	Name                string             `bson:"name" json:"name"`
	Comment             string             `bson:"comment" json:"comment"`
	InitDiskSize        int                `bson:"init_disk_size" json:"init_disk_size"`
	Memory              int                `bson:"memory" json:"memory"`
	Processors          int                `bson:"processors" json:"processors"`
	NetworkRoles        []string           `bson:"network_roles" json:"network_roles"`
	Isos                []*iso.Iso         `bson:"isos" json:"isos"`
	UsbDevices          []*usb.Device      `bson:"usb_devices" json:"usb_devices"`
	PciDevices          []*pci.Device      `bson:"pci_devices" json:"pci_devices"`
	DriveDevices        []*drive.Device    `bson:"drive_devices" json:"drive_devices"`
	IscsiDevices        []*iscsi.Device    `bson:"iscsi_devices" json:"iscsi_devices"`
	Vnc                 bool               `bson:"vnc" json:"vnc"`
	VncPassword         string             `bson:"vnc_password" json:"vnc_password"`
	VncDisplay          int                `bson:"vnc_display,omitempty" json:"vnc_display"`
	Virt                *vm.VirtualMachine `bson:"-" json:"-"`
	curVpc              primitive.ObjectID `bson:"-" json:"-"`
	curSubnet           primitive.ObjectID `bson:"-" json:"-"`
	curDeleteProtection bool               `bson:"-" json:"-"`
	curState            string             `bson:"-" json:"-"`
	curNoPublicAddress  bool               `bson:"-" json:"-"`
	curNoHostAddress    bool               `bson:"-" json:"-"`
}

func (i *Instance) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	if i.State == "" {
		i.State = Start
	}

	if i.State != Start {
		i.Restart = false
		i.RestartBlockIp = false
	}

	if !ValidStates.Contains(i.State) {
		errData = &errortypes.ErrorData{
			Error:   "invalid_state",
			Message: "Invalid instance state",
		}
		return
	}

	if i.Organization.IsZero() {
		errData = &errortypes.ErrorData{
			Error:   "organization_required",
			Message: "Missing required organization",
		}
		return
	}

	if i.Zone.IsZero() {
		errData = &errortypes.ErrorData{
			Error:   "zone_required",
			Message: "Missing required zone",
		}
		return
	}

	if i.Node.IsZero() {
		errData = &errortypes.ErrorData{
			Error:   "node_required",
			Message: "Missing required node",
		}
		return
	}

	if i.Vpc.IsZero() {
		errData = &errortypes.ErrorData{
			Error:   "vpc_required",
			Message: "Missing required VPC",
		}
		return
	}

	vc, err := vpc.Get(db, i.Vpc)
	if err != nil {
		return
	}

	if i.Subnet.IsZero() {
		errData = &errortypes.ErrorData{
			Error:   "vpc_subnet_required",
			Message: "Missing required VPC subnet",
		}
		return
	}

	sub := vc.GetSubnet(i.Subnet)
	if sub == nil {
		errData = &errortypes.ErrorData{
			Error:   "vpc_subnet_missing",
			Message: "VPC subnet does not exist",
		}
		return
	}

	if i.InitDiskSize != 0 && i.InitDiskSize < 10 {
		errData = &errortypes.ErrorData{
			Error:   "init_disk_size_invalid",
			Message: "Disk size below minimum",
		}
		return
	}

	if i.Memory < 256 {
		i.Memory = 256
	}

	if i.Processors < 1 {
		i.Processors = 1
	}

	if i.NetworkRoles == nil {
		i.NetworkRoles = []string{}
	}

	if i.PublicIps == nil {
		i.PublicIps = []string{}
	}

	if i.PublicIps6 == nil {
		i.PublicIps6 = []string{}
	}

	if i.PrivateIps == nil {
		i.PrivateIps = []string{}
	}

	if i.PrivateIps6 == nil {
		i.PrivateIps6 = []string{}
	}

	nde, err := node.Get(db, i.Node)
	if err != nil {
		return
	}

	if i.OracleSubnet != "" {
		match := false
		for _, subnet := range nde.OracleSubnets {
			if subnet == i.OracleSubnet {
				match = true
				break
			}
		}

		if !match {
			errData = &errortypes.ErrorData{
				Error:   "oracle_subnet_invalid",
				Message: "Invalid Oracle subnet",
			}
			return
		}
	}

	if i.Isos == nil {
		i.Isos = []*iso.Iso{}
	} else {
		for _, is := range i.Isos {
			is.Name = utils.FilterPath(is.Name, 128)
		}
	}

	if i.UsbDevices == nil {
		i.UsbDevices = []*usb.Device{}
	} else {
		for _, device := range i.UsbDevices {
			device.Name = ""
			device.Vendor = usb.FilterId(device.Vendor)
			device.Product = usb.FilterId(device.Product)
			device.Bus = usb.FilterAddr(device.Bus)
			device.Address = usb.FilterAddr(device.Address)

			if (device.Vendor == "" || device.Product == "") &&
				(device.Bus == "" || device.Address == "") {

				errData = &errortypes.ErrorData{
					Error:   "usb_device_invalid",
					Message: "Invalid USB device",
				}
				return
			}
		}
	}

	if i.PciDevices == nil {
		i.PciDevices = []*pci.Device{}
	} else {
		for _, device := range i.PciDevices {
			device.Name = ""
			device.Class = ""
			device.Driver = ""

			if !pci.CheckSlot(device.Slot) {
				errData = &errortypes.ErrorData{
					Error:   "pci_device_slot_invalid",
					Message: "Invalid PCI slot",
				}
				return
			}
		}
	}

	instanceDrives := set.NewSet()
	nodeInstanceDrives := nde.InstanceDrives
	if nodeInstanceDrives != nil {
		for _, device := range nodeInstanceDrives {
			instanceDrives.Add(device.Id)
		}
	}

	if i.DriveDevices == nil {
		i.DriveDevices = []*drive.Device{}
	} else {
		for _, device := range i.DriveDevices {
			if !instanceDrives.Contains(device.Id) {
				errData = &errortypes.ErrorData{
					Error:   "drive_invalid",
					Message: "Instance drive not available",
				}
				return
			}
		}
	}

	iscsiDevices := []*iscsi.Device{}
	if i.IscsiDevices != nil {
		for _, device := range i.IscsiDevices {
			if device.Uri == "" {
				continue
			}

			errData, err = device.Parse()
			if err != nil || errData != nil {
				return
			}

			iscsiDevices = append(iscsiDevices, device)
		}
	}
	i.IscsiDevices = iscsiDevices

	if i.Vnc {
		if i.VncDisplay == 0 {
			i.VncDisplay = rand.Intn(9998) + 4101
		}
		if i.VncPassword == "" {
			i.VncPassword, err = utils.RandStr(32)
			if err != nil {
				return
			}
		}
	} else {
		i.VncPassword = ""
	}

	return
}

func (i *Instance) Format() {
	// TODO Sort VPC IDs
}

func (i *Instance) Json() {
	switch i.State {
	case Start:
		if i.Restart || i.RestartBlockIp {
			i.Status = "Restart Required"
		} else {
			switch i.VmState {
			case vm.Starting:
				i.Status = "Starting"
				break
			case vm.Running:
				i.Status = "Running"
				break
			case vm.Stopped:
				i.Status = "Starting"
				break
			case vm.Failed:
				i.Status = "Starting"
				break
			case vm.Updating:
				i.Status = "Updating"
				break
			case vm.Provisioning:
				i.Status = "Provisioning"
				break
			case "":
				i.Status = "Provisioning"
				break
			}
		}
		break
	case Cleanup:
		switch i.VmState {
		case vm.Starting:
			i.Status = "Stopping"
			break
		case vm.Running:
			i.Status = "Stopping"
			break
		case vm.Stopped:
			i.Status = "Stopping"
			break
		case vm.Failed:
			i.Status = "Stopping"
			break
		case vm.Updating:
			i.Status = "Updating"
			break
		case vm.Provisioning:
			i.Status = "Stopping"
			break
		case "":
			i.Status = "Stopping"
			break
		}
		break
	case Stop:
		switch i.VmState {
		case vm.Starting:
			i.Status = "Stopping"
			break
		case vm.Running:
			i.Status = "Stopping"
			break
		case vm.Stopped:
			i.Status = "Stopped"
			break
		case vm.Failed:
			i.Status = "Failed"
			break
		case vm.Updating:
			i.Status = "Updating"
			break
		case vm.Provisioning:
			i.Status = "Stopped"
			break
		case "":
			i.Status = "Stopped"
			break
		}
		break
	case Restart:
		i.Status = "Restarting"
		break
	case Destroy:
		i.Status = "Destroying"
		break
	}

	i.PublicMac = vm.GetMacAddrExternal(i.Id, i.Vpc)
	if i.VmTimestamp.IsZero() {
		i.Uptime = ""
	} else {
		i.Uptime = systemd.FormatUptime(i.VmTimestamp)
	}

	if i.IscsiDevices != nil {
		for _, device := range i.IscsiDevices {
			device.Json()
		}
	}
}

func (i *Instance) IsActive() bool {
	return i.State == Start || i.VmState == vm.Running ||
		i.VmState == vm.Starting || i.VmState == vm.Provisioning
}

func (i *Instance) PreCommit() {
	i.curVpc = i.Vpc
	i.curSubnet = i.Subnet
	i.curDeleteProtection = i.DeleteProtection
	i.curState = i.State
	i.curNoPublicAddress = i.NoPublicAddress
	i.curNoHostAddress = i.NoHostAddress
}

func (i *Instance) PostCommit(db *database.Database) (
	dskChange bool, err error) {

	if (!i.curVpc.IsZero() && i.curVpc != i.Vpc) ||
		(!i.curSubnet.IsZero() && i.curSubnet != i.Subnet) {

		err = vpc.RemoveInstanceIp(db, i.Id, i.curVpc)
		if err != nil {
			return
		}
	}

	if i.curDeleteProtection != i.DeleteProtection {
		dskChange = true

		err = disk.SetDeleteProtection(db, i.Id, i.DeleteProtection)
		if err != nil {
			return
		}
	}

	if i.curState != i.State && (i.State == Stop || i.State == Start ||
		i.State == Restart) {

		i.Restart = false
		i.RestartBlockIp = false
	}

	if i.curNoPublicAddress != i.NoPublicAddress && i.NoPublicAddress {
		err = block.RemoveInstanceIpsType(db, i.Id, block.External)
		if err != nil {
			return
		}
	}

	if i.curNoHostAddress != i.NoHostAddress && i.NoHostAddress {
		err = block.RemoveInstanceIpsType(db, i.Id, block.Host)
		if err != nil {
			return
		}
	}

	return
}

func (i *Instance) Commit(db *database.Database) (err error) {
	coll := db.Instances()

	err = coll.Commit(i.Id, i)
	if err != nil {
		return
	}

	return
}

func (i *Instance) CommitFields(db *database.Database, fields set.Set) (
	err error) {

	coll := db.Instances()

	err = coll.CommitFields(i.Id, i, fields)
	if err != nil {
		return
	}

	return
}

func (i *Instance) Insert(db *database.Database) (err error) {
	coll := db.Instances()

	if !i.Id.IsZero() {
		err = &errortypes.DatabaseError{
			errors.New("instance: Instance already exists"),
		}
		return
	}

	_, err = coll.InsertOne(db, i)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func (i *Instance) LoadVirt(disks []*disk.Disk) {
	i.Virt = &vm.VirtualMachine{
		Id:         i.Id,
		Image:      i.Image,
		Processors: i.Processors,
		Memory:     i.Memory,
		Hugepages:  node.Self.Hugepages,
		Vnc:        i.Vnc,
		VncDisplay: i.VncDisplay,
		Disks:      []*vm.Disk{},
		NetworkAdapters: []*vm.NetworkAdapter{
			&vm.NetworkAdapter{
				Type:       vm.Bridge,
				MacAddress: vm.GetMacAddr(i.Id, i.Vpc),
				Vpc:        i.Vpc,
				Subnet:     i.Subnet,
			},
		},
		OracleSubnet:     i.OracleSubnet,
		OracleVnic:       i.OracleVnic,
		OracleVnicAttach: i.OracleVnicAttach,
		Uefi:             i.Uefi,
		SecureBoot:       i.SecureBoot,
		NoPublicAddress:  i.NoPublicAddress,
		NoHostAddress:    i.NoHostAddress,
		Isos:             []*vm.Iso{},
		UsbDevices:       []*vm.UsbDevice{},
		PciDevices:       []*vm.PciDevice{},
		DriveDevices:     []*vm.DriveDevice{},
		IscsiDevices:     []*vm.IscsiDevice{},
	}

	if disks != nil {
		for _, dsk := range disks {
			index, err := strconv.Atoi(dsk.Index)
			if err != nil {
				continue
			}

			i.Virt.Disks = append(i.Virt.Disks, &vm.Disk{
				Id:    dsk.Id,
				Index: index,
				Path:  paths.GetDiskPath(dsk.Id),
			})
		}
	}

	for _, is := range i.Isos {
		i.Virt.Isos = append(i.Virt.Isos, &vm.Iso{
			Name: is.Name,
		})
	}

	if node.Self.UsbPassthrough && i.UsbDevices != nil {
		for _, device := range i.UsbDevices {
			i.Virt.UsbDevices = append(i.Virt.UsbDevices, &vm.UsbDevice{
				Vendor:  device.Vendor,
				Product: device.Product,
				Bus:     device.Bus,
				Address: device.Address,
			})
		}
	}

	if node.Self.PciPassthrough && i.PciDevices != nil {
		for _, device := range i.PciDevices {
			i.Virt.PciDevices = append(i.Virt.PciDevices, &vm.PciDevice{
				Slot: device.Slot,
			})
		}
	}

	instanceDrives := set.NewSet()
	nodeInstanceDrives := node.Self.InstanceDrives
	if nodeInstanceDrives != nil {
		for _, device := range nodeInstanceDrives {
			instanceDrives.Add(device.Id)
		}
	}

	if i.DriveDevices != nil {
		for _, device := range i.DriveDevices {
			if instanceDrives.Contains(device.Id) {
				i.Virt.DriveDevices = append(
					i.Virt.DriveDevices,
					&vm.DriveDevice{
						Id: device.Id,
					},
				)
			}
		}
	}

	if node.Self.Iscsi && i.IscsiDevices != nil {
		for _, device := range i.IscsiDevices {
			i.Virt.IscsiDevices = append(
				i.Virt.IscsiDevices,
				&vm.IscsiDevice{
					Uri: device.QemuUri(),
				},
			)
		}
	}

	return
}

func (i *Instance) Changed(curVirt *vm.VirtualMachine) bool {
	if i.Virt.Memory != curVirt.Memory ||
		i.Virt.Hugepages != curVirt.Hugepages ||
		i.Virt.Processors != curVirt.Processors ||
		i.Virt.Vnc != curVirt.Vnc ||
		i.Virt.VncDisplay != curVirt.VncDisplay ||
		i.Virt.Uefi != curVirt.Uefi ||
		i.Virt.SecureBoot != curVirt.SecureBoot ||
		i.Virt.NoPublicAddress != curVirt.NoPublicAddress ||
		i.Virt.NoHostAddress != curVirt.NoHostAddress {

		return true
	}

	for i, adapter := range i.Virt.NetworkAdapters {
		if len(curVirt.NetworkAdapters) <= i {
			return true
		}

		if adapter.Vpc != curVirt.NetworkAdapters[i].Vpc {
			return true
		}

		if adapter.Subnet != curVirt.NetworkAdapters[i].Subnet {
			return true
		}
	}

	if i.Virt.Isos != nil {
		if len(i.Virt.Isos) > 0 && curVirt.Isos == nil {
			return true
		}

		for i, device := range i.Virt.Isos {
			if len(curVirt.Isos) <= i {
				return true
			}

			if device.Name != curVirt.Isos[i].Name {
				return true
			}
		}
	}

	if i.Virt.PciDevices != nil {
		if len(i.Virt.PciDevices) > 0 && curVirt.PciDevices == nil {
			return true
		}

		for i, device := range i.Virt.PciDevices {
			if len(curVirt.PciDevices) <= i {
				return true
			}

			if device.Slot != curVirt.PciDevices[i].Slot {
				return true
			}
		}
	}

	if i.Virt.DriveDevices != nil {
		if len(i.Virt.DriveDevices) > 0 && curVirt.DriveDevices == nil {
			return true
		}

		for i, device := range i.Virt.DriveDevices {
			if len(curVirt.DriveDevices) <= i {
				return true
			}

			if device.Id != curVirt.DriveDevices[i].Id {
				return true
			}
		}
	}

	if i.Virt.IscsiDevices != nil {
		if len(i.Virt.IscsiDevices) > 0 && curVirt.IscsiDevices == nil {
			return true
		}

		for i, device := range i.Virt.IscsiDevices {
			if len(curVirt.IscsiDevices) <= i {
				return true
			}

			if device.Uri != curVirt.IscsiDevices[i].Uri {
				return true
			}
		}
	}

	return false
}

func (i *Instance) DiskChanged(curVirt *vm.VirtualMachine) (
	addDisks, remDisks []*vm.Disk) {

	addDisks = []*vm.Disk{}
	remDisks = []*vm.Disk{}

	if !curVirt.DisksAvailable {
		logrus.WithFields(logrus.Fields{
			"instance_id": curVirt.Id.Hex(),
		}).Warn("qemu: Ignoring disk state")
		return
	}

	disks := set.NewSet()
	curDisks := set.NewSet()

	for _, dsk := range i.Virt.Disks {
		disks.Add(dsk.Id)
	}

	for _, dsk := range curVirt.Disks {
		if !disks.Contains(dsk.Id) {
			remDisks = append(remDisks, dsk)
		} else {
			curDisks.Add(dsk.Id)
		}
	}

	for _, dsk := range i.Virt.Disks {
		if !curDisks.Contains(dsk.Id) {
			addDisks = append(addDisks, dsk)
		}
	}

	return
}

func (i *Instance) UsbChanged(curVirt *vm.VirtualMachine) (
	addUsbs, remUsbs []*vm.UsbDevice) {

	addUsbs = []*vm.UsbDevice{}
	remUsbs = []*vm.UsbDevice{}

	if !node.Self.UsbPassthrough {
		return
	}

	if !curVirt.UsbDevicesAvailable {
		logrus.WithFields(logrus.Fields{
			"instance_id": curVirt.Id.Hex(),
		}).Warn("qemu: Ignoring USB state")
		return
	}

	usbsVendor := set.NewSet()
	curUsbsVendor := set.NewSet()
	usbsBus := set.NewSet()
	curUsbsBus := set.NewSet()

	if curVirt.UsbDevices != nil {
		for _, device := range curVirt.UsbDevices {
			if device.Vendor != "" && device.Product != "" {
				curUsbsVendor.Add(fmt.Sprintf("%s_%s",
					device.Vendor, device.Product))
			} else if device.Bus != "" && device.Address != "" {
				curUsbsBus.Add(fmt.Sprintf("%s_%s",
					device.Bus, device.Address))
			}
		}
	}

	if i.Virt.UsbDevices != nil {
		for _, device := range i.Virt.UsbDevices {
			if device.Vendor != "" && device.Product != "" {
				usbsVendor.Add(fmt.Sprintf("%s_%s",
					device.Vendor, device.Product))
			} else if device.Bus != "" && device.Address != "" {
				usbsBus.Add(fmt.Sprintf("%s_%s",
					device.Bus, device.Address))
			}
		}
	}

	addUsbsVendor := usbsVendor.Copy()
	addUsbsVendor.Subtract(curUsbsVendor)
	addUsbsBus := usbsBus.Copy()
	addUsbsBus.Subtract(curUsbsBus)
	remUsbsVendor := curUsbsVendor.Copy()
	remUsbsVendor.Subtract(usbsVendor)
	remUsbsBus := curUsbsBus.Copy()
	remUsbsBus.Subtract(usbsBus)

	for deviceInf := range addUsbsVendor.Iter() {
		device := strings.Split(deviceInf.(string), "_")
		addUsbs = append(addUsbs, &vm.UsbDevice{
			Vendor:  device[0],
			Product: device[1],
		})
	}
	for deviceInf := range addUsbsBus.Iter() {
		device := strings.Split(deviceInf.(string), "_")
		addUsbs = append(addUsbs, &vm.UsbDevice{
			Bus:     device[0],
			Address: device[1],
		})
	}
	for deviceInf := range remUsbsVendor.Iter() {
		device := strings.Split(deviceInf.(string), "_")
		remUsbs = append(remUsbs, &vm.UsbDevice{
			Vendor:  device[0],
			Product: device[1],
		})
	}
	for deviceInf := range remUsbsBus.Iter() {
		device := strings.Split(deviceInf.(string), "_")
		remUsbs = append(remUsbs, &vm.UsbDevice{
			Bus:     device[0],
			Address: device[1],
		})
	}

	return
}

func (i *Instance) VncConnect(db *database.Database,
	rw http.ResponseWriter, r *http.Request) (err error) {

	nde, err := node.Get(db, i.Node)
	if err != nil {
		return
	}

	if nde.PublicIps == nil || len(nde.PublicIps) == 0 {
		err = &errortypes.NotFoundError{
			errors.New("instance: Node missing public IP for VNC"),
		}
		return
	}

	wsUrl := fmt.Sprintf(
		"ws://%s:%d",
		nde.PublicIps[0],
		i.VncDisplay+15900,
	)

	var backConn *websocket.Conn
	var backResp *http.Response

	dialer := &websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	header := http.Header{}
	header.Set(
		"Sec-Websocket-Protocol",
		r.Header.Get("Sec-Websocket-Protocol"),
	)

	backConn, backResp, err = dialer.Dial(wsUrl, header)
	if err != nil {
		if backResp != nil {
			err = &VncDialError{
				errors.Wrapf(err, "instance: WebSocket dial error %d",
					backResp.StatusCode),
			}
		} else {
			err = &VncDialError{
				errors.Wrap(err, "instance: WebSocket dial error"),
			}
		}
		return
	}
	defer backConn.Close()

	wsUpgrader := &websocket.Upgrader{
		HandshakeTimeout: time.Duration(
			settings.Router.HandshakeTimeout) * time.Second,
		ReadBufferSize:  2048,
		WriteBufferSize: 2048,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	upgradeHeader := http.Header{}
	val := backResp.Header.Get("Sec-Websocket-Protocol")
	if val != "" {
		upgradeHeader.Set("Sec-Websocket-Protocol", val)
	}

	frontConn, err := wsUpgrader.Upgrade(rw, r, upgradeHeader)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "instance: WebSocket upgrade error"),
		}
		return
	}
	defer frontConn.Close()

	wait := make(chan bool, 4)
	go func() {
		defer func() {
			rec := recover()
			if rec != nil {
				logrus.WithFields(logrus.Fields{
					"panic": rec,
				}).Error("instance: WebSocket VNC back panic")
				wait <- true
			}
		}()
		io.Copy(backConn.UnderlyingConn(), frontConn.UnderlyingConn())
		wait <- true
	}()
	go func() {
		defer func() {
			rec := recover()
			if rec != nil {
				logrus.WithFields(logrus.Fields{
					"panic": rec,
				}).Error("instance: WebSocket VNC front panic")
				wait <- true
			}
		}()
		io.Copy(frontConn.UnderlyingConn(), backConn.UnderlyingConn())
		wait <- true
	}()
	<-wait

	return
}

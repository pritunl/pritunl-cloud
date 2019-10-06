package instance

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/block"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/systemd"
	"github.com/pritunl/pritunl-cloud/usb"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/pritunl/pritunl-cloud/vpc"
)

type Instance struct {
	Id                  primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Organization        primitive.ObjectID `bson:"organization" json:"organization"`
	Zone                primitive.ObjectID `bson:"zone" json:"zone"`
	Vpc                 primitive.ObjectID `bson:"vpc" json:"vpc"`
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
	DeleteProtection    bool               `bson:"delete_protection" json:"delete_protection"`
	PublicIps           []string           `bson:"public_ips" json:"public_ips"`
	PublicIps6          []string           `bson:"public_ips6" json:"public_ips6"`
	PrivateIps          []string           `bson:"private_ips" json:"private_ips"`
	PrivateIps6         []string           `bson:"private_ips6" json:"private_ips6"`
	HostIps             []string           `bson:"host_ips" json:"host_ips"`
	NoPublicAddress     bool               `bson:"no_public_address" json:"no_public_address"`
	NoHostAddress       bool               `bson:"no_host_address" json:"no_host_address"`
	Node                primitive.ObjectID `bson:"node" json:"node"`
	Domain              primitive.ObjectID `bson:"domain,omitempty" json:"domain"`
	Name                string             `bson:"name" json:"name"`
	InitDiskSize        int                `bson:"init_disk_size" json:"init_disk_size"`
	Memory              int                `bson:"memory" json:"memory"`
	Processors          int                `bson:"processors" json:"processors"`
	NetworkRoles        []string           `bson:"network_roles" json:"network_roles"`
	UsbDevices          []*usb.Device      `bson:"usb_devices" json:"usb_devices"`
	Vnc                 bool               `bson:"vnc" json:"vnc"`
	VncPassword         string             `bson:"vnc_password" json:"vnc_password"`
	VncDisplay          int                `bson:"vnc_display,omitempty" json:"vnc_display"`
	Virt                *vm.VirtualMachine `bson:"-" json:"-"`
	curVpc              primitive.ObjectID `bson:"-" json:"-"`
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

	if i.Image.IsZero() {
		errData = &errortypes.ErrorData{
			Error:   "image_required",
			Message: "Missing required image",
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

	if i.UsbDevices == nil {
		i.UsbDevices = []*usb.Device{}
	} else {
		for _, device := range i.UsbDevices {
			device.Name = ""
			device.Vendor = usb.FilterId(device.Vendor)
			device.Product = usb.FilterId(device.Product)

			if device.Vendor == "" || device.Product == "" {
				errData = &errortypes.ErrorData{
					Error:   "usb_device_invalid",
					Message: "Invalid USB device",
				}
				return
			}
		}
	}

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
}

func (i *Instance) IsActive() bool {
	return i.State == Start || i.VmState == vm.Running ||
		i.VmState == vm.Starting || i.VmState == vm.Provisioning
}

func (i *Instance) PreCommit() {
	i.curVpc = i.Vpc
	i.curDeleteProtection = i.DeleteProtection
	i.curState = i.State
	i.curNoPublicAddress = i.NoPublicAddress
	i.curNoHostAddress = i.NoHostAddress
}

func (i *Instance) PostCommit(db *database.Database) (
	dskChange bool, err error) {

	if !i.curVpc.IsZero() && i.curVpc != i.Vpc {
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
		Vnc:        i.Vnc,
		VncDisplay: i.VncDisplay,
		Disks:      []*vm.Disk{},
		NetworkAdapters: []*vm.NetworkAdapter{
			&vm.NetworkAdapter{
				Type:       vm.Bridge,
				MacAddress: vm.GetMacAddr(i.Id, i.Vpc),
				VpcId:      i.Vpc,
			},
		},
		NoPublicAddress: i.NoPublicAddress,
		NoHostAddress:   i.NoHostAddress,
	}

	if disks != nil {
		for _, dsk := range disks {
			index, err := strconv.Atoi(dsk.Index)
			if err != nil {
				continue
			}

			i.Virt.Disks = append(i.Virt.Disks, &vm.Disk{
				Index: index,
				Path:  paths.GetDiskPath(dsk.Id),
			})
		}
	}

	return
}

func (i *Instance) Changed(curVirt *vm.VirtualMachine) bool {
	if i.Virt.Memory != curVirt.Memory ||
		i.Virt.Processors != curVirt.Processors ||
		i.Virt.Vnc != curVirt.Vnc ||
		i.Virt.VncDisplay != curVirt.VncDisplay ||
		i.Virt.NoPublicAddress != curVirt.NoPublicAddress ||
		i.Virt.NoHostAddress != curVirt.NoHostAddress {

		return true
	}

	for i, adapter := range i.Virt.NetworkAdapters {
		if len(curVirt.NetworkAdapters) <= i {
			return true
		}

		if adapter.VpcId != curVirt.NetworkAdapters[i].VpcId {
			return true
		}
	}

	return false
}

func (i *Instance) DiskChanged(curVirt *vm.VirtualMachine) (
	addDisks, remDisks []*vm.Disk) {

	addDisks = []*vm.Disk{}
	remDisks = []*vm.Disk{}
	disks := set.NewSet()
	curDisks := map[int]*vm.Disk{}

	for _, dsk := range i.Virt.Disks {
		disks.Add(dsk.Index)
	}

	for _, dsk := range curVirt.Disks {
		if !disks.Contains(dsk.Index) {
			remDisks = append(remDisks, dsk)
		} else {
			curDisks[dsk.Index] = dsk
		}
	}

	for _, dsk := range i.Virt.Disks {
		curDsk := curDisks[dsk.Index]
		if curDsk == nil {
			addDisks = append(addDisks, dsk)
		} else if dsk.Path != curDsk.Path {
			remDisks = append(remDisks, curDsk)
			addDisks = append(addDisks, dsk)
		}
	}

	return
}

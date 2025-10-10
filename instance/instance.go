package instance

import (
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/gorilla/websocket"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/block"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/drive"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/iscsi"
	"github.com/pritunl/pritunl-cloud/iso"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/nodeport"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/pci"
	"github.com/pritunl/pritunl-cloud/pool"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/shape"
	"github.com/pritunl/pritunl-cloud/systemd"
	"github.com/pritunl/pritunl-cloud/telemetry"
	"github.com/pritunl/pritunl-cloud/tpm"
	"github.com/pritunl/pritunl-cloud/usb"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/pritunl/pritunl-cloud/vpc"
	"github.com/sirupsen/logrus"
)

var scriptReg = regexp.MustCompile("^#!")

type Instance struct {
	Id                  bson.ObjectID       `bson:"_id,omitempty" json:"id"`
	Organization        bson.ObjectID       `bson:"organization" json:"organization"`
	UnixId              int                 `bson:"unix_id" json:"unix_id"`
	Datacenter          bson.ObjectID       `bson:"datacenter" json:"datacenter"`
	Zone                bson.ObjectID       `bson:"zone" json:"zone"`
	Vpc                 bson.ObjectID       `bson:"vpc" json:"vpc"`
	Subnet              bson.ObjectID       `bson:"subnet" json:"subnet"`
	Created             time.Time           `bson:"created" json:"created"`
	Guest               *GuestData          `bson:"guest,omitempty" json:"guest"`
	CloudSubnet         string              `bson:"cloud_subnet" json:"cloud_subnet"`
	CloudVnic           string              `bson:"cloud_vnic" json:"cloud_vnic"`
	CloudVnicAttach     string              `bson:"cloud_vnic_attach" json:"cloud_vnic_attach"`
	Image               bson.ObjectID       `bson:"image" json:"image"`
	ImageBacking        bool                `bson:"image_backing" json:"image_backing"`
	DiskType            string              `bson:"disk_type" json:"disk_type"`
	DiskPool            bson.ObjectID       `bson:"disk_pool" json:"disk_pool"`
	Status              string              `bson:"-" json:"status"`
	StatusInfo          *StatusInfo         `bson:"status_info,omitempty" json:"status_info"`
	Uptime              string              `bson:"-" json:"uptime"`
	State               string              `bson:"state" json:"state"`
	Action              string              `bson:"action" json:"action"`
	PublicMac           string              `bson:"-" json:"public_mac"`
	Timestamp           time.Time           `bson:"timestamp" json:"timestamp"`
	Restart             bool                `bson:"restart" json:"restart"`
	RestartReason       string              `bson:"restart_reason" json:"restart_reason"`
	RestartBlockIp      bool                `bson:"restart_block_ip" json:"restart_block_ip"`
	Uefi                bool                `bson:"uefi" json:"uefi"`
	SecureBoot          bool                `bson:"secure_boot" json:"secure_boot"`
	Tpm                 bool                `bson:"tpm" json:"tpm"`
	TpmSecret           string              `bson:"tpm_secret" json:"-"`
	DhcpServer          bool                `bson:"dhcp_server" json:"dhcp_server"`
	CloudType           string              `bson:"cloud_type" json:"cloud_type"`
	CloudScript         string              `bson:"cloud_script" json:"cloud_script"`
	DeleteProtection    bool                `bson:"delete_protection" json:"delete_protection"`
	SkipSourceDestCheck bool                `bson:"skip_source_dest_check" json:"skip_source_dest_check"`
	QemuVersion         string              `bson:"qemu_version" json:"qemu_version"`
	PublicIps           []string            `bson:"public_ips" json:"public_ips"`
	PublicIps6          []string            `bson:"public_ips6" json:"public_ips6"`
	PrivateIps          []string            `bson:"private_ips" json:"private_ips"`
	PrivateIps6         []string            `bson:"private_ips6" json:"private_ips6"`
	GatewayIps          []string            `bson:"gateway_ips" json:"gateway_ips"`
	GatewayIps6         []string            `bson:"gateway_ips6" json:"gateway_ips6"`
	CloudPrivateIps     []string            `bson:"cloud_private_ips" json:"cloud_private_ips"`
	CloudPublicIps      []string            `bson:"cloud_public_ips" json:"cloud_public_ips"`
	CloudPublicIps6     []string            `bson:"cloud_public_ips6" json:"cloud_public_ips6"`
	HostIps             []string            `bson:"host_ips" json:"host_ips"`
	NodePortIps         []string            `bson:"node_port_ips" json:"node_port_ips"`
	NodePorts           []*nodeport.Mapping `bson:"node_ports,omitempty" json:"node_ports"`
	DhcpIp              string              `bson:"dhcp_ip" json:"dhcp_ip"`
	DhcpIp6             string              `bson:"dhcp_ip6" json:"dhcp_ip6"`
	NetworkNamespace    string              `bson:"network_namespace" json:"network_namespace"`
	NoPublicAddress     bool                `bson:"no_public_address" json:"no_public_address"`
	NoPublicAddress6    bool                `bson:"no_public_address6" json:"no_public_address6"`
	NoHostAddress       bool                `bson:"no_host_address" json:"no_host_address"`
	Node                bson.ObjectID       `bson:"node" json:"node"`
	Shape               bson.ObjectID       `bson:"shape" json:"shape"`
	Name                string              `bson:"name" json:"name"`
	Comment             string              `bson:"comment" json:"comment"`
	RootEnabled         bool                `bson:"root_enabled" json:"root_enabled"`
	RootPasswd          string              `bson:"root_passwd" json:"root_passwd"`
	InitDiskSize        int                 `bson:"init_disk_size" json:"init_disk_size"`
	Memory              int                 `bson:"memory" json:"memory"`
	Processors          int                 `bson:"processors" json:"processors"`
	Roles               []string            `bson:"roles" json:"roles"`
	Isos                []*iso.Iso          `bson:"isos,omitempty" json:"isos"`
	UsbDevices          []*usb.Device       `bson:"usb_devices,omitempty" json:"usb_devices"`
	PciDevices          []*pci.Device       `bson:"pci_devices,omitempty" json:"pci_devices"`
	DriveDevices        []*drive.Device     `bson:"drive_devices,omitempty" json:"drive_devices"`
	IscsiDevices        []*iscsi.Device     `bson:"iscsi_devices,omitempty" json:"iscsi_devices"`
	Mounts              []*Mount            `bson:"mounts,omitempty" json:"mounts"`
	Vnc                 bool                `bson:"vnc" json:"vnc"`
	VncPassword         string              `bson:"vnc_password" json:"vnc_password"`
	VncDisplay          int                 `bson:"vnc_display" json:"vnc_display"`
	Spice               bool                `bson:"spice" json:"spice"`
	SpicePassword       string              `bson:"spice_password" json:"spice_password"`
	SpicePort           int                 `bson:"spice_port" json:"spice_port"`
	Gui                 bool                `bson:"gui" json:"gui"`
	Deployment          bson.ObjectID       `bson:"deployment" json:"deployment"`
	Info                *Info               `bson:"info,omitempty" json:"info"`
	Virt                *vm.VirtualMachine  `bson:"-" json:"-"`

	curVpc              bson.ObjectID                       `bson:"-" json:"-"`
	curSubnet           bson.ObjectID                       `bson:"-" json:"-"`
	curDeleteProtection bool                                `bson:"-" json:"-"`
	curAction           string                              `bson:"-" json:"-"`
	curNoPublicAddress  bool                                `bson:"-" json:"-"`
	curNoHostAddress    bool                                `bson:"-" json:"-"`
	curNodePorts        map[bson.ObjectID]*nodeport.Mapping `bson:"-" json:"-"`
	removedNodePorts    []bson.ObjectID                     `bson:"-" json:"-"`
	newId               bool                                `bson:"-" json:"-"`
}

type Completion struct {
	Id           bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name         string        `bson:"name" json:"name"`
	Organization bson.ObjectID `bson:"organization" json:"organization"`
	Zone         bson.ObjectID `bson:"zone" json:"zone"`
	Vpc          bson.ObjectID `bson:"vpc" json:"vpc"`
	Subnet       bson.ObjectID `bson:"subnet" json:"subnet"`
	Node         bson.ObjectID `bson:"node" json:"node"`
}

type Mount struct {
	Name     string `bson:"name" json:"name"`
	Type     string `bson:"type" json:"type"`
	Path     string `bson:"path" json:"path"`
	HostPath string `bson:"host_path" json:"host_path"`
}

type StatusInfo struct {
	DownloadProgress int     `bson:"download_progress,omitempty" json:"download_progress"`
	DownloadSpeed    float64 `bson:"download_speed,omitempty" json:"download_speed"`
}

type GuestData struct {
	Status    string              `bson:"status" json:"status"`
	Timestamp time.Time           `bson:"timestamp" json:"timestamp"`
	Heartbeat time.Time           `bson:"heartbeat" json:"heartbeat"`
	Memory    float64             `bson:"memory" json:"memory"`
	HugePages float64             `bson:"hugepages" json:"hugepages"`
	Load1     float64             `bson:"load1" json:"load1"`
	Load5     float64             `bson:"load5" json:"load5"`
	Load15    float64             `bson:"load15" json:"load15"`
	Updates   []*telemetry.Update `bson:"updates" json:"updates"`
}

type Info struct {
	Node          string              `bson:"node" json:"node"`
	NodePublicIp  string              `bson:"node_public_ip" json:"node_public_ip"`
	Mtu           int                 `bson:"mtu" json:"mtu"`
	Iscsi         bool                `bson:"iscsi" json:"iscsi"`
	Disks         []string            `bson:"disks" json:"disks"`
	FirewallRules map[string]string   `bson:"firewall_rules" json:"firewall_rules"`
	Authorities   []string            `bson:"authorities" json:"authorities"`
	Isos          []*iso.Iso          `bson:"isos" json:"isos"`
	UsbDevices    []*usb.Device       `bson:"usb_devices" json:"usb_devices"`
	PciDevices    []*pci.Device       `bson:"pci_devices" json:"pci_devices"`
	DriveDevices  []*drive.Device     `bson:"drive_devices" json:"drive_devices"`
	CloudSubnets  []*node.CloudSubnet `bson:"cloud_subnets" json:"cloud_subnets"`
	Timestamp     time.Time           `bson:"timestamp" json:"timestamp"`
}

func (i *Instance) GenerateId() (err error) {
	if !i.Id.IsZero() {
		err = &errortypes.DatabaseError{
			errors.New("instance: Instance already exists"),
		}
		return
	}

	i.newId = true
	i.Id = bson.NewObjectID()

	return
}

func (i *Instance) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	i.Name = utils.FilterName(i.Name)

	if i.Action == "" {
		i.Action = Start
	}

	if i.Action != Start {
		i.Restart = false
		i.RestartReason = ""
		i.RestartBlockIp = false
	}

	if i.State != "" && !ValidStates.Contains(i.State) {
		errData = &errortypes.ErrorData{
			Error:   "invalid_state",
			Message: "Invalid instance state",
		}
		return
	}

	if !ValidActions.Contains(i.Action) {
		errData = &errortypes.ErrorData{
			Error:   "invalid_action",
			Message: "Invalid instance action",
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
		if i.Shape.IsZero() {
			errData = &errortypes.ErrorData{
				Error:   "node_required",
				Message: "Missing required node",
			}
			return
		}

		shpe, e := shape.Get(db, i.Shape)
		if e != nil {
			err = e
			return
		}

		if !shpe.Flexible {
			i.Processors = shpe.Processors
			i.Memory = shpe.Memory
		}

		nde, e := shpe.FindNode(db, i.Processors, i.Memory)
		if e != nil {
			err = e
			return
		}

		i.Node = nde.Id
		i.DiskType = shpe.DiskType
		i.DiskPool = shpe.DiskPool
	}

	if i.Vpc.IsZero() {
		errData = &errortypes.ErrorData{
			Error:   "vpc_required",
			Message: "Missing required VPC",
		}
		return
	}

	if i.UnixId == 0 {
		i.GenerateUnixId()
	}

	switch i.DiskType {
	case disk.Lvm:
		if i.DiskPool.IsZero() {
			errData = &errortypes.ErrorData{
				Error:   "pool_required",
				Message: "Missing required disk pool",
			}
			return
		}
		break
	case disk.Qcow2, "":
		i.DiskType = disk.Qcow2
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

	if i.Roles == nil {
		i.Roles = []string{}
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

	if i.CloudType == "" {
		i.CloudType = Linux
	}
	if !ValidCloudTypes.Contains(i.CloudType) {
		errData = &errortypes.ErrorData{
			Error:   "invalid_cloud_type",
			Message: "Invalid cloud init type",
		}
		return
	}

	if i.CloudScript != "" && !scriptReg.MatchString(i.CloudScript) {
		errData = &errortypes.ErrorData{
			Error:   "invalid_cloud_script",
			Message: "Startup script missing shebang on first line",
		}
		return
	}

	if i.TpmSecret == "" {
		i.TpmSecret, err = tpm.GenerateSecret()
		if err != nil {
			return
		}
	}

	nde, err := node.Get(db, i.Node)
	if err != nil {
		return
	}

	if i.Datacenter == bson.NilObjectID {
		i.Datacenter = nde.Datacenter
	}

	if i.Datacenter != vc.Datacenter {
		errData = &errortypes.ErrorData{
			Error:   "vpc_invalid_datacenter",
			Message: "VPC must be in same datacenter as instance",
		}
		return
	}

	if i.CloudSubnet != "" {
		match := false
		for _, subnet := range nde.CloudSubnets {
			if subnet == i.CloudSubnet {
				match = true
				break
			}
		}

		if !match {
			errData = &errortypes.ErrorData{
				Error:   "cloud_subnet_invalid",
				Message: "Invalid Cloud subnet",
			}
			return
		}
	}

	if i.RootEnabled {
		if i.RootPasswd == "" {
			i.RootPasswd, err = utils.RandPasswd(8)
			if err != nil {
				return
			}
		}
	} else {
		i.RootPasswd = ""
	}

	if i.Isos == nil {
		i.Isos = []*iso.Iso{}
	} else {
		for _, is := range i.Isos {
			is.Name = utils.FilterRelPath(is.Name)
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

			available, e := usb.Available(db, i.Id, i.Node, device)
			if e != nil {
				err = e
				return
			}

			if !available {
				errData = &errortypes.ErrorData{
					Error:   "usb_device_unavailable",
					Message: "USB device in use by another instance",
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
	for _, device := range nodeInstanceDrives {
		instanceDrives.Add(device.Id)
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

	newMounts := []*Mount{}
	for _, mount := range i.Mounts {
		if mount.Name == "" && mount.HostPath == "" {
			continue
		}

		mount.Name = utils.FilterNameCmd(mount.Name)
		mount.Type = HostPath
		mount.Path = utils.FilterPath(mount.Path)
		mount.HostPath = utils.FilterPath(mount.HostPath)

		if mount.Name == "" {
			errData = &errortypes.ErrorData{
				Error:   "missing_mount_name",
				Message: "Missing required mount name",
			}
			return
		}

		if mount.HostPath == "" {
			errData = &errortypes.ErrorData{
				Error:   "mount_host_path_invalid",
				Message: "Mount host path invalid",
			}
			return
		}

		newMounts = append(newMounts, mount)
	}
	i.Mounts = newMounts

	if i.Vnc {
		if i.VncPassword == "" {
			i.VncPassword, err = utils.RandPasswd(32)
			if err != nil {
				return
			}
		}
	} else {
		i.VncPassword = ""
	}

	if i.Spice {
		if i.SpicePassword == "" {
			i.SpicePassword, err = utils.RandPasswd(32)
			if err != nil {
				return
			}
		}
	} else {
		i.SpicePassword = ""
	}

	externalNodePorts := set.NewSet()
	for _, mapping := range i.NodePorts {
		extPortKey := fmt.Sprintf("%s:%d",
			mapping.Protocol, mapping.ExternalPort)

		if !mapping.Delete {
			if externalNodePorts.Contains(extPortKey) {
				errData = &errortypes.ErrorData{
					Error:   "node_port_external_duplicate",
					Message: "Duplicate external node port",
				}
				return
			}
		}
		externalNodePorts.Add(extPortKey)

		errData, err = mapping.Validate(db)
		if err != nil {
			return
		}

		available, e := nodeport.Available(db, i.Datacenter, i.Organization,
			mapping.Protocol, mapping.ExternalPort)
		if e != nil {
			err = e
			return
		}

		if !available {
			errData = &errortypes.ErrorData{
				Error:   "node_port_unavailable",
				Message: "External node port is unavailable",
			}
			return
		}
	}

	return
}

func (i *Instance) GenerateUnixId() {
	i.UnixId = rand.Intn(55500) + 10000
}

func (i *Instance) InitUnixId(db *database.Database) (err error) {
	if i.UnixId != 0 {
		return
	}

	i.GenerateUnixId()

	err = i.CommitFields(db, set.NewSet("unix_id"))
	if err != nil {
		return
	}

	return
}

func (i *Instance) GenerateSpicePort() {
	// Spice 15000 - 19999
	i.SpicePort = rand.Intn(4999) + 15000
}

func (i *Instance) InitSpicePort(db *database.Database) (err error) {
	if i.SpicePort != 0 {
		return
	}

	i.GenerateSpicePort()

	coll := db.Instances()

	for n := 0; n < 10000; n++ {
		err = coll.CommitFields(i.Id, i, set.NewSet("spice_port"))
		if err != nil {
			err = database.ParseError(err)
			if _, ok := err.(*database.DuplicateKeyError); ok {
				i.GenerateSpicePort()
				err = nil
				continue
			}
			return
		}

		event.PublishDispatch(db, "instance.change")

		return
	}

	err = &errortypes.WriteError{
		errors.New("instance: Failed to commit unique spice port"),
	}
	return
}

func (i *Instance) GenerateVncDisplay() {
	// VNC 10001 - 14999 (+5900)
	// VNC WebSocket 20001 - 24999 (+15900)
	i.VncDisplay = rand.Intn(4999) + 4101
}

func (i *Instance) InitVncDisplay(db *database.Database) (err error) {
	if i.VncDisplay != 0 {
		return
	}

	i.GenerateVncDisplay()

	coll := db.Instances()

	for n := 0; n < 10000; n++ {
		err = coll.CommitFields(i.Id, i, set.NewSet("vnc_display"))
		if err != nil {
			err = database.ParseError(err)
			if _, ok := err.(*database.DuplicateKeyError); ok {
				i.GenerateVncDisplay()
				err = nil
				continue
			}
			return
		}

		event.PublishDispatch(db, "instance.change")

		return
	}

	err = &errortypes.WriteError{
		errors.New("instance: Failed to commit unique vnc port"),
	}
	return
}

func (i *Instance) Format() {
}

func (i *Instance) Json(short bool) {
	switch i.Action {
	case Start:
		if i.Restart || i.RestartBlockIp {
			i.Status = "Restart Required"
			if i.RestartReason != "" {
				i.Status += fmt.Sprintf(" (%s)", i.RestartReason)
			}
		} else {
			switch i.State {
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
		switch i.State {
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
		switch i.State {
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

	if !i.IsActive() && i.Guest != nil {
		i.Guest.Timestamp = time.Time{}
		i.Guest.Heartbeat = time.Time{}
		i.Guest.Memory = 0
		i.Guest.HugePages = 0
		i.Guest.Load1 = 0
		i.Guest.Load5 = 0
		i.Guest.Load15 = 0
	}

	i.PublicMac = vm.GetMacAddrExternal(i.Id, i.Vpc)
	if i.Timestamp.IsZero() || !i.IsActive() {
		i.Uptime = ""
	} else {
		if short {
			i.Uptime = systemd.FormatUptimeShort(i.Timestamp)
		} else {
			i.Uptime = systemd.FormatUptime(i.Timestamp)
		}
	}

	if i.IscsiDevices != nil {
		for _, device := range i.IscsiDevices {
			device.Json()
		}
	}
}

func (i *Instance) IsActive() bool {
	return i.Action == Start || i.State == vm.Running ||
		i.State == vm.Starting || i.State == vm.Provisioning
}

func (i *Instance) IsIpv6Only() bool {
	return (node.Self.NetworkMode == node.Disabled || i.NoPublicAddress) &&
		(node.Self.NetworkMode6 != node.Disabled && !i.NoPublicAddress6) &&
		(node.Self.NoHostNetwork || i.NoHostAddress)
}

func (i *Instance) PreCommit() {
	i.curVpc = i.Vpc
	i.curSubnet = i.Subnet
	i.curDeleteProtection = i.DeleteProtection
	i.curAction = i.Action
	i.curNoPublicAddress = i.NoPublicAddress
	i.curNoHostAddress = i.NoHostAddress

	nodePortMap := map[bson.ObjectID]*nodeport.Mapping{}
	for _, mapping := range i.NodePorts {
		nodePortMap[mapping.NodePort] = mapping
	}
	i.curNodePorts = nodePortMap
}

func (i *Instance) UpsertNodePorts(newNodePorts []*nodeport.Mapping) {
	if len(i.NodePorts) == 0 {
		i.NodePorts = newNodePorts
		return
	}

	processed := make(map[int]bool)
	newMappings := []*nodeport.Mapping{}

	for _, newMapping := range newNodePorts {
		matched := false

		if newMapping.ExternalPort != 0 {
			for x, curMapping := range i.NodePorts {
				if curMapping.Protocol == newMapping.Protocol &&
					curMapping.InternalPort == newMapping.InternalPort &&
					curMapping.ExternalPort == newMapping.ExternalPort {

					newMapping.NodePort = curMapping.NodePort
					newMappings = append(newMappings, newMapping)

					processed[x] = true
					matched = true
					break
				}
			}
		} else {
			for x, curMapping := range i.NodePorts {
				if curMapping.Protocol == newMapping.Protocol &&
					curMapping.InternalPort == newMapping.InternalPort &&
					!processed[x] {

					newMapping.NodePort = curMapping.NodePort
					newMapping.ExternalPort = curMapping.ExternalPort
					newMappings = append(newMappings, newMapping)

					processed[x] = true
					matched = true
					break
				}
			}
		}

		if !matched {
			newMappings = append(newMappings, newMapping)
		}
	}

	i.NodePorts = newMappings
}

func (i *Instance) SyncNodePorts(db *database.Database) (err error) {
	newNodePorts := []*nodeport.Mapping{}
	newNodePortIds := set.NewSet()
	externalPorts := set.NewSet()

	if i.curNodePorts == nil {
		i.curNodePorts = map[bson.ObjectID]*nodeport.Mapping{}
	}

	for _, mapping := range i.NodePorts {
		if !mapping.NodePort.IsZero() {
			curMapping := i.curNodePorts[mapping.NodePort]
			if curMapping == nil {
				continue
			}
			newNodePortIds.Add(curMapping.NodePort)

			if mapping.Delete {
				i.removedNodePorts = append(
					i.removedNodePorts, curMapping.NodePort)
				continue
			}

			curMapping.InternalPort = mapping.InternalPort
			mapping = curMapping
		}

		var errData *errortypes.ErrorData
		var ndePort *nodeport.NodePort
		if mapping.ExternalPort != 0 {
			ndePort, err = nodeport.GetPort(db, i.Datacenter, i.Organization,
				mapping.Protocol, mapping.ExternalPort)
			if err != nil {
				if _, ok := err.(*database.NotFoundError); ok {
					ndePort = nil
					err = nil
				} else {
					return
				}
			}
		}

		if ndePort == nil {
			ndePort, errData, err = nodeport.New(db,
				i.Datacenter, i.Organization,
				mapping.Protocol, mapping.ExternalPort)
			if err != nil {
				return
			}
			if errData != nil {
				err = errData.GetError()
				return
			}
		}

		mapping.NodePort = ndePort.Id
		mapping.ExternalPort = ndePort.Port

		extPortKey := fmt.Sprintf("%s:%d",
			mapping.Protocol, mapping.ExternalPort)
		if externalPorts.Contains(extPortKey) {
			continue
		}
		externalPorts.Add(extPortKey)

		newNodePorts = append(newNodePorts, mapping)
	}
	i.NodePorts = newNodePorts

	for _, mapping := range i.curNodePorts {
		if newNodePortIds.Contains(mapping.NodePort) {
			continue
		}
		i.removedNodePorts = append(i.removedNodePorts, mapping.NodePort)
	}

	return
}

func (i *Instance) PostCommit(db *database.Database) (
	dskChange bool, err error) {

	err = i.SyncNodePorts(db)
	if err != nil {
		return
	}

	if (!i.curVpc.IsZero() && i.curVpc != i.Vpc) ||
		(!i.curSubnet.IsZero() && i.curSubnet != i.Subnet) {

		i.DhcpIp = ""
		i.DhcpIp6 = ""

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

	if i.curAction != i.Action && (i.Action == Stop || i.Action == Start ||
		i.Action == Restart) {

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

func (i *Instance) Cleanup(db *database.Database) (err error) {
	for _, mapping := range i.NodePorts {
		ndePort, e := nodeport.Get(db, mapping.NodePort)
		if e != nil {
			err = e
			if _, ok := err.(*database.NotFoundError); ok {
				err = nil
				continue
			}
			return
		}

		err = ndePort.Sync(db)
		if err != nil {
			return
		}
	}

	for _, ndePortId := range i.removedNodePorts {
		ndePort, e := nodeport.Get(db, ndePortId)
		if e != nil {
			err = e
			if _, ok := err.(*database.NotFoundError); ok {
				err = nil
				continue
			}
			return
		}

		err = ndePort.Sync(db)
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

	if fields.Contains("unix_id") {
		for n := 0; n < 10000; n++ {
			err = coll.CommitFields(i.Id, i, fields)
			if err != nil {
				err = database.ParseError(err)
				if _, ok := err.(*database.DuplicateKeyError); ok {
					i.GenerateUnixId()
					err = nil
					continue
				}
				return
			}

			return
		}

		err = &errortypes.WriteError{
			errors.New("instance: Failed to commit unique unix id"),
		}
		return
	} else {
		err = coll.CommitFields(i.Id, i, fields)
		if err != nil {
			return
		}
	}

	return
}

func (i *Instance) Insert(db *database.Database) (err error) {
	coll := db.Instances()

	if !i.Id.IsZero() && !i.newId {
		err = &errortypes.DatabaseError{
			errors.New("instance: Instance already exists"),
		}
		return
	}

	i.Created = time.Now()

	for n := 0; n < 2000; n++ {
		resp, e := coll.InsertOne(db, i)
		if e != nil {
			err = database.ParseError(e)
			if _, ok := err.(*database.DuplicateKeyError); ok {
				i.GenerateUnixId()
				err = nil
				continue
			}
			return
		}

		i.Id = resp.InsertedID.(bson.ObjectID)
		return
	}

	err = &errortypes.WriteError{
		errors.New("instance: Failed to insert unique unix id"),
	}
	return
}

func (i *Instance) LoadVirt(poolsMap map[bson.ObjectID]*pool.Pool,
	disks []*disk.Disk) {

	i.Virt = &vm.VirtualMachine{
		Id:           i.Id,
		Organization: i.Organization,
		UnixId:       i.UnixId,
		DiskType:     i.DiskType,
		DiskPool:     i.DiskPool,
		Image:        i.Image,
		Processors:   i.Processors,
		Memory:       i.Memory,
		Hugepages:    node.Self.Hugepages,
		Vnc:          i.Vnc,
		VncDisplay:   i.VncDisplay,
		Spice:        i.Spice,
		SpicePort:    i.SpicePort,
		Gui:          i.Gui,
		Disks:        []*vm.Disk{},
		NetworkAdapters: []*vm.NetworkAdapter{
			&vm.NetworkAdapter{
				Type:       vm.Bridge,
				MacAddress: vm.GetMacAddr(i.Id, i.Vpc),
				Vpc:        i.Vpc,
				Subnet:     i.Subnet,
			},
		},
		CloudSubnet:      i.CloudSubnet,
		CloudVnic:        i.CloudVnic,
		CloudVnicAttach:  i.CloudVnicAttach,
		DhcpIp:           i.DhcpIp,
		DhcpIp6:          i.DhcpIp6,
		Uefi:             i.Uefi,
		SecureBoot:       i.SecureBoot,
		Tpm:              i.Tpm,
		DhcpServer:       i.DhcpServer,
		Deployment:       i.Deployment,
		CloudType:        i.CloudType,
		NoPublicAddress:  i.NoPublicAddress,
		NoPublicAddress6: i.NoPublicAddress6,
		NoHostAddress:    i.NoHostAddress,
		Isos:             []*vm.Iso{},
		UsbDevices:       []*vm.UsbDevice{},
		PciDevices:       []*vm.PciDevice{},
		DriveDevices:     []*vm.DriveDevice{},
		IscsiDevices:     []*vm.IscsiDevice{},
		Mounts:           []*vm.Mount{},
	}

	if disks != nil {
		for _, dsk := range disks {
			switch dsk.Type {
			case disk.Lvm:
				if poolsMap == nil {
					continue
				}

				pl := poolsMap[dsk.Pool]
				if pl == nil {
					continue
				}

				i.Virt.DriveDevices = append(
					i.Virt.DriveDevices,
					&vm.DriveDevice{
						Id:     dsk.Id.Hex(),
						Type:   vm.Lvm,
						VgName: pl.VgName,
						LvName: dsk.Id.Hex(),
					},
				)
				break
			case disk.Qcow2, "":
				index, err := strconv.Atoi(dsk.Index)
				if err != nil {
					continue
				}

				i.Virt.Disks = append(i.Virt.Disks, &vm.Disk{
					Id:    dsk.Id,
					Index: index,
					Path:  paths.GetDiskPath(dsk.Id),
				})
				break
			}
		}
	}

	for _, is := range i.Isos {
		i.Virt.Isos = append(i.Virt.Isos, &vm.Iso{
			Name: is.Name,
		})
	}

	if node.Self.UsbPassthrough && i.UsbDevices != nil {
		for _, device := range i.UsbDevices {
			usbDevice, _ := usb.GetDevice(
				device.Bus, device.Address,
				device.Vendor, device.Product,
			)

			if usbDevice != nil {
				i.Virt.UsbDevices = append(i.Virt.UsbDevices, &vm.UsbDevice{
					Vendor:  usbDevice.Vendor,
					Product: usbDevice.Product,
					Bus:     usbDevice.Bus,
					Address: usbDevice.Address,
				})
			}
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
						Id:   device.Id,
						Type: vm.Physical,
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

	for _, mount := range i.Mounts {
		i.Virt.Mounts = append(
			i.Virt.Mounts,
			&vm.Mount{
				Name:     mount.Name,
				Type:     mount.Type,
				Path:     mount.Path,
				HostPath: mount.HostPath,
			},
		)
	}

	return
}

func (i *Instance) Changed(curVirt *vm.VirtualMachine) (bool, string) {
	curCloudType := curVirt.CloudType
	if curCloudType == "" {
		curCloudType = Linux
	}
	cloudType := i.Virt.CloudType
	if cloudType == "" {
		cloudType = Linux
	}

	if i.Virt.Memory != curVirt.Memory {
		return true, "Memory size changed"
	}
	if i.Virt.Hugepages != curVirt.Hugepages {
		return true, "Hugepages changed"
	}
	if i.Virt.Processors != curVirt.Processors {
		return true, "Processor count changed"
	}
	if i.Virt.Vnc != curVirt.Vnc {
		return true, "VNC changed"
	}
	if i.Virt.VncDisplay != curVirt.VncDisplay {
		return true, "VNC display changed"
	}
	if i.Virt.Spice != curVirt.Spice {
		return true, "SPICE changed"
	}
	if i.Virt.SpicePort != curVirt.SpicePort {
		return true, "SPICE port changed"
	}
	if i.Virt.Gui != curVirt.Gui {
		return true, "GUI changed"
	}
	if i.Virt.Uefi != curVirt.Uefi {
		return true, "UEFI changed"
	}
	if i.Virt.SecureBoot != curVirt.SecureBoot {
		return true, "Secure boot changed"
	}
	if i.Virt.Tpm != curVirt.Tpm {
		return true, "TPM changed"
	}
	if i.Virt.DhcpServer != curVirt.DhcpServer {
		return true, "DHCP server changed"
	}
	if cloudType != curCloudType {
		return true, "Cloud type changed"
	}
	if i.Virt.NoPublicAddress != curVirt.NoPublicAddress {
		return true, "Public address changed"
	}
	if i.Virt.NoPublicAddress6 != curVirt.NoPublicAddress6 {
		return true, "Public IPv6 changed"
	}
	if i.Virt.NoHostAddress != curVirt.NoHostAddress {
		return true, "Host address changed"
	}

	for i, adapter := range i.Virt.NetworkAdapters {
		if len(curVirt.NetworkAdapters) <= i {
			return true, "Network adapters changed"
		}

		if adapter.Vpc != curVirt.NetworkAdapters[i].Vpc {
			return true, "VPC changed"
		}

		if adapter.Subnet != curVirt.NetworkAdapters[i].Subnet {
			return true, "Subnet changed"
		}
	}

	if i.Virt.Isos != nil {
		if len(i.Virt.Isos) > 0 && curVirt.Isos == nil {
			return true, "ISO devices changed"
		}

		for i, device := range i.Virt.Isos {
			if len(curVirt.Isos) <= i {
				return true, "ISO devices changed"
			}

			if device.Name != curVirt.Isos[i].Name {
				return true, "ISO device changed"
			}
		}
	}

	if i.Virt.PciDevices != nil {
		if len(i.Virt.PciDevices) > 0 && curVirt.PciDevices == nil {
			return true, "PCI devices changed"
		}

		for i, device := range i.Virt.PciDevices {
			if len(curVirt.PciDevices) <= i {
				return true, "PCI devices changed"
			}

			if device.Slot != curVirt.PciDevices[i].Slot {
				return true, "PCI device slot changed"
			}
		}
	}

	if i.Virt.DriveDevices != nil {
		if len(i.Virt.DriveDevices) > 0 && curVirt.DriveDevices == nil {
			return true, "Drive devices changed"
		}

		for i, device := range i.Virt.DriveDevices {
			if len(curVirt.DriveDevices) <= i {
				return true, "Drive devices changed"
			}

			if device.Id != curVirt.DriveDevices[i].Id {
				return true, "Drive device changed"
			}
		}
	}

	if i.Virt.IscsiDevices != nil {
		if len(i.Virt.IscsiDevices) > 0 && curVirt.IscsiDevices == nil {
			return true, "iSCSI devices changed"
		}

		for i, device := range i.Virt.IscsiDevices {
			if len(curVirt.IscsiDevices) <= i {
				return true, "iSCSI devices changed"
			}

			if device.Uri != curVirt.IscsiDevices[i].Uri {
				return true, "iSCSI URI changed"
			}
		}
	}

	if i.Virt.Mounts != nil {
		if len(i.Virt.Mounts) > 0 && curVirt.Mounts == nil {
			return true, "Mounts changed"
		}

		for i, mount := range i.Virt.Mounts {
			if len(curVirt.Mounts) <= i {
				return true, "Mounts changed"
			}

			if mount.Name != curVirt.Mounts[i].Name {
				return true, "Mount name changed"
			}
			if mount.Type != curVirt.Mounts[i].Type {
				return true, "Mount type changed"
			}
			if mount.Path != curVirt.Mounts[i].Path {
				return true, "Mount path changed"
			}
			if mount.HostPath != curVirt.Mounts[i].HostPath {
				return true, "Mount host path changed"
			}
		}
	}

	return false, ""
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

	disks := map[bson.ObjectID]*vm.Disk{}
	curDisks := set.NewSet()

	for _, dsk := range i.Virt.Disks {
		disks[dsk.Id] = dsk
	}

	for _, dsk := range curVirt.Disks {
		newDsk := disks[dsk.Id]

		if newDsk == nil || dsk.Index != newDsk.Index {
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

	usbs := set.NewSet()
	usbsMap := map[string]*vm.UsbDevice{}
	curUsbs := set.NewSet()
	curUsbsMap := map[string]*vm.UsbDevice{}

	if curVirt.UsbDevices != nil {
		for _, device := range curVirt.UsbDevices {
			key := device.Key()
			curUsbs.Add(key)
			curUsbsMap[key] = device
		}
	}

	if i.Virt.UsbDevices != nil {
		for _, device := range i.Virt.UsbDevices {
			key := device.Key()
			usbs.Add(key)
			usbsMap[key] = device
		}
	}

	addUsbsSet := usbs.Copy()
	addUsbsSet.Subtract(curUsbs)
	remUsbsSet := curUsbs.Copy()
	remUsbsSet.Subtract(usbs)

	for deviceInf := range addUsbsSet.Iter() {
		device := usbsMap[deviceInf.(string)]
		addUsbs = append(addUsbs, device)
	}
	for deviceInf := range remUsbsSet.Iter() {
		device := curUsbsMap[deviceInf.(string)]
		remUsbs = append(remUsbs, device)
	}

	return
}

func (i *Instance) VncConnect(db *database.Database,
	rw http.ResponseWriter, r *http.Request) (err error) {

	nde, err := node.Get(db, i.Node)
	if err != nil {
		return
	}

	vncHost := ""
	if nde.Id == node.Self.Id {
		vncHost = "127.0.0.1"
	} else if nde.PublicIps == nil || len(nde.PublicIps) == 0 {
		err = &errortypes.NotFoundError{
			errors.New("instance: Node missing public IP for VNC"),
		}
		return
	} else {
		vncHost = nde.PublicIps[0]
	}

	wsUrl := fmt.Sprintf(
		"ws://%s:%d",
		vncHost,
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

		for {
			msgType, msg, err := frontConn.ReadMessage()
			if err != nil {
				closeMsg := websocket.FormatCloseMessage(
					websocket.CloseNormalClosure, fmt.Sprintf("%v", err))
				if e, ok := err.(*websocket.CloseError); ok {
					if e.Code != websocket.CloseNoStatusReceived {
						closeMsg = websocket.FormatCloseMessage(e.Code, e.Text)
					}
				}
				_ = backConn.WriteMessage(websocket.CloseMessage, closeMsg)
				break
			}

			err = backConn.WriteMessage(msgType, msg)
			if err != nil {
				err = &errortypes.ReadError{
					errors.Wrap(err, "instance: WebSocket VNC write error"),
				}
				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Error("instance: WebSocket VNC back write error")
				break
			}
		}

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

		for {
			msgType, msg, err := backConn.ReadMessage()
			if err != nil {
				closeMsg := websocket.FormatCloseMessage(
					websocket.CloseNormalClosure, fmt.Sprintf("%v", err))
				if e, ok := err.(*websocket.CloseError); ok {
					if e.Code != websocket.CloseNoStatusReceived {
						closeMsg = websocket.FormatCloseMessage(e.Code, e.Text)
					}
				}
				_ = frontConn.WriteMessage(websocket.CloseMessage, closeMsg)
				break
			}

			err = frontConn.WriteMessage(msgType, msg)
			if err != nil {
				err = &errortypes.ReadError{
					errors.Wrap(err, "instance: WebSocket VNC write error"),
				}
				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Error("instance: WebSocket VNC back write error")
				break
			}
		}

		wait <- true
	}()
	<-wait

	return
}

package vm

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/usb"
	"github.com/pritunl/pritunl-cloud/utils"
)

type VirtualMachine struct {
	Id                  primitive.ObjectID `json:"id"`
	Organization        primitive.ObjectID `json:"organization"`
	UnixId              int                `json:"unix_id"`
	State               string             `json:"state"`
	Timestamp           time.Time          `json:"timestamp"`
	QemuVersion         string             `json:"qemu_version"`
	DiskType            string             `json:"disk_type"`
	DiskPool            primitive.ObjectID `json:"disk_pool"`
	Image               primitive.ObjectID `json:"image"`
	Processors          int                `json:"processors"`
	Memory              int                `json:"memory"`
	Hugepages           bool               `json:"hugepages"`
	Vnc                 bool               `json:"vnc"`
	VncDisplay          int                `json:"vnc_display"`
	Spice               bool               `json:"spice"`
	SpicePort           int                `json:"spice_port"`
	Gui                 bool               `json:"gui"`
	Disks               []*Disk            `json:"disks"`
	DisksAvailable      bool               `json:"-"`
	NetworkAdapters     []*NetworkAdapter  `json:"network_adapters"`
	OracleSubnet        string             `json:"oracle_subnet"`
	OracleVnic          string             `json:"oracle_vnic"`
	OracleVnicAttach    string             `json:"oracle_vnic_attach"`
	OraclePrivateIp     string             `json:"oracle_private_ip"`
	OraclePublicIp      string             `json:"oracle_public_ip"`
	OraclePublicIp6     string             `json:"oracle_public_ip6"`
	Uefi                bool               `json:"uefi"`
	SecureBoot          bool               `json:"secure_boot"`
	Tpm                 bool               `json:"tpm"`
	DhcpServer          bool               `json:"dhcp_server"`
	Deployment          primitive.ObjectID `json:"deployment"`
	CloudType           string             `json:"cloud_type"`
	NoPublicAddress     bool               `json:"no_public_address"`
	NoPublicAddress6    bool               `json:"no_public_address6"`
	NoHostAddress       bool               `json:"no_host_address"`
	Isos                []*Iso             `json:"isos"`
	UsbDevices          []*UsbDevice       `json:"usb_devices"`
	UsbDevicesAvailable bool               `json:"-"`
	PciDevices          []*PciDevice       `json:"pci_devices"`
	DriveDevices        []*DriveDevice     `json:"drive_devices"`
	IscsiDevices        []*IscsiDevice     `json:"iscsi_devices"`
	Mounts              []*Mount           `json:"mounts"`
	ImdsVersion         int                `json:"imds_version"`
	ImdsClientSecret    string             `json:"-"`
	ImdsHostSecret      string             `json:"imds_host_secret"`
}

func (v *VirtualMachine) HasExternalNetwork() bool {
	return v.Vnc || v.Spice || (v.IscsiDevices != nil &&
		len(v.IscsiDevices) > 0)
}

func (v *VirtualMachine) ProtectHome() bool {
	return !v.Gui
}

func (v *VirtualMachine) ProtectTmp() bool {
	return !v.Gui
}

func (v *VirtualMachine) Running() bool {
	return v.State == Starting || v.State == Running
}

func (v *VirtualMachine) GenerateImdsSecret() (err error) {
	v.ImdsVersion = 1

	v.ImdsClientSecret, err = utils.RandStr(32)
	if err != nil {
		return
	}

	v.ImdsHostSecret, err = utils.RandStr(32)
	if err != nil {
		return
	}

	return
}

type Disk struct {
	Id    primitive.ObjectID `json:"id"`
	Index int                `json:"index"`
	Path  string             `json:"path"`
}

type Iso struct {
	Name string `json:"name"`
}

type UsbDevice struct {
	Id      string `json:"id"`
	Vendor  string `json:"vendor"`
	Product string `json:"product"`
	Bus     string `json:"bus"`
	Address string `json:"address"`
}

func (u *UsbDevice) Key() string {
	return fmt.Sprintf("%s_%s_%s_%s",
		u.Bus,
		u.Address,
		u.Vendor,
		u.Product,
	)
}

func (u *UsbDevice) Copy() (device *UsbDevice) {
	device = &UsbDevice{
		Id:      u.Id,
		Vendor:  u.Vendor,
		Product: u.Product,
		Bus:     u.Bus,
		Address: u.Address,
	}

	return
}

func (u *UsbDevice) GetQemuId() string {
	return fmt.Sprintf("usbd_%s_%s_%s_%s_%d",
		u.Bus,
		u.Address,
		u.Vendor,
		u.Product,
		utils.RandInt(1111, 9999),
	)
}

func (u *UsbDevice) GetDevice() (device *usb.Device, err error) {
	device, err = usb.GetDevice(u.Bus, u.Address, u.Vendor, u.Product)
	if err != nil {
		return
	}

	return
}

type PciDevice struct {
	Slot string `json:"slot"`
}

type DriveDevice struct {
	Id     string `json:"id"`
	Type   string `json:"type"`
	VgName string `json:"vg_name"`
	LvName string `json:"lv_name"`
}

type IscsiDevice struct {
	Uri string `json:"iscsi"`
}

type Mount struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Path     string `json:"path"`
	HostPath string `json:"host_path"`
}

func (d *Disk) GetId() primitive.ObjectID {
	idStr := strings.Split(path.Base(d.Path), ".")[0]

	objId, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return primitive.NilObjectID
	}
	return objId
}

func (d *Disk) Copy() (dsk *Disk) {
	dsk = &Disk{
		Id:    d.Id,
		Index: d.Index,
		Path:  d.Path,
	}

	return
}

type NetworkAdapter struct {
	Type       string             `json:"type"`
	MacAddress string             `json:"mac_address"`
	Vpc        primitive.ObjectID `json:"vpc"`
	Subnet     primitive.ObjectID `json:"subnet"`
	IpAddress  string             `json:"ip_address,omitempty"`
	IpAddress6 string             `json:"ip_address6,omitempty"`
}

func (v *VirtualMachine) Commit(db *database.Database) (err error) {
	coll := db.Instances()

	addrs := []string{}
	addrs6 := []string{}

	for _, adapter := range v.NetworkAdapters {
		if adapter.IpAddress != "" {
			addrs = append(addrs, adapter.IpAddress)
		}
		if adapter.IpAddress6 != "" {
			addrs6 = append(addrs6, adapter.IpAddress6)
		}
	}

	data := bson.M{
		"virt_state":     v.State,
		"virt_timestamp": v.Timestamp,
		"public_ips":     addrs,
		"public_ips6":    addrs6,
	}

	if v.QemuVersion != "" {
		data["qemu_version"] = v.QemuVersion
	}

	err = coll.UpdateId(v.Id, &bson.M{
		"$set": data,
	})
	if err != nil {
		err = database.ParseError(err)
		if _, ok := err.(*database.NotFoundError); ok {
			err = nil
		} else {
			return
		}
	}

	if !v.Deployment.IsZero() {
		coll = db.Deployments()

		err = coll.UpdateId(v.Deployment, &bson.M{
			"$set": &bson.M{
				"instance_data.public_ips":  addrs,
				"instance_data.public_ips6": addrs6,
			},
		})
		if err != nil {
			err = database.ParseError(err)
			return
		}
	}

	return
}

func (v *VirtualMachine) CommitOracleVnic(db *database.Database) (err error) {
	coll := db.Instances()

	err = coll.UpdateId(v.Id, &bson.M{
		"$set": &bson.M{
			"oracle_vnic":        v.OracleVnic,
			"oracle_vnic_attach": v.OracleVnicAttach,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func (v *VirtualMachine) CommitOracleIps(db *database.Database) (err error) {
	coll := db.Instances()

	oraclePivateAddrs := []string{}
	if v.OraclePrivateIp != "" {
		oraclePivateAddrs = append(oraclePivateAddrs, v.OraclePrivateIp)
	}

	oraclePublicAddrs := []string{}
	if v.OraclePublicIp != "" {
		oraclePublicAddrs = append(oraclePublicAddrs, v.OraclePublicIp)
	}

	oraclePublicAddrs6 := []string{}
	if v.OraclePublicIp6 != "" {
		oraclePublicAddrs6 = append(oraclePublicAddrs6, v.OraclePublicIp6)
	}

	err = coll.UpdateId(v.Id, &bson.M{
		"$set": &bson.M{
			"oracle_private_ips": oraclePivateAddrs,
			"oracle_public_ips":  oraclePublicAddrs,
			"oracle_public_ips6": oraclePublicAddrs6,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	if !v.Deployment.IsZero() {
		coll = db.Deployments()

		err = coll.UpdateId(v.Deployment, &bson.M{
			"$set": &bson.M{
				"instance_data.oracle_private_ips": oraclePivateAddrs,
				"instance_data.oracle_public_ips":  oraclePublicAddrs,
				"instance_data.oracle_public_ips6": oraclePublicAddrs6,
			},
		})
		if err != nil {
			err = database.ParseError(err)
			return
		}
	}

	return
}

func (v *VirtualMachine) CommitState(db *database.Database, action string) (
	err error) {

	coll := db.Instances()

	addrs := []string{}
	addrs6 := []string{}

	for _, adapter := range v.NetworkAdapters {
		if adapter.IpAddress != "" {
			addrs = append(addrs, adapter.IpAddress)
		}
		if adapter.IpAddress6 != "" {
			addrs6 = append(addrs6, adapter.IpAddress6)
		}
	}

	data := bson.M{
		"action":         action,
		"virt_state":     v.State,
		"virt_timestamp": v.Timestamp,
		"public_ips":     addrs,
		"public_ips6":    addrs6,
	}

	if v.QemuVersion != "" {
		data["qemu_version"] = v.QemuVersion
	}

	err = coll.UpdateId(v.Id, &bson.M{
		"$set": data,
	})
	if err != nil {
		err = database.ParseError(err)
		if _, ok := err.(*database.NotFoundError); ok {
			err = nil
		} else {
			return
		}
	}

	return
}

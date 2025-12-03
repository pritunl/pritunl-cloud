package types

import (
	"time"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/instance"
)

type Instance struct {
	Id                  bson.ObjectID `json:"id"`
	Organization        bson.ObjectID `json:"organization"`
	Zone                bson.ObjectID `json:"zone"`
	Vpc                 bson.ObjectID `json:"vpc"`
	Subnet              bson.ObjectID `json:"subnet"`
	CloudSubnet         string        `json:"cloud_subnet"`
	CloudVnic           string        `json:"cloud_vnic"`
	Image               bson.ObjectID `json:"image"`
	State               string        `json:"state"`
	Timestamp           time.Time     `json:"timestamp"`
	Action              string        `json:"action"`
	Uefi                bool          `json:"uefi"`
	SecureBoot          bool          `json:"secure_boot"`
	Tpm                 bool          `json:"tpm"`
	DhcpServer          bool          `json:"dhcp_server"`
	CloudType           string        `json:"cloud_type"`
	SystemKind          string        `json:"system_kind"`
	DeleteProtection    bool          `json:"delete_protection"`
	SkipSourceDestCheck bool          `json:"skip_source_dest_check"`
	QemuVersion         string        `json:"qemu_version"`
	PublicIps           []string      `json:"public_ips"`
	PublicIps6          []string      `json:"public_ips6"`
	PrivateIps          []string      `json:"private_ips"`
	PrivateIps6         []string      `json:"private_ips6"`
	GatewayIps          []string      `json:"gateway_ips"`
	GatewayIps6         []string      `json:"gateway_ips6"`
	CloudPrivateIps     []string      `json:"cloud_private_ips"`
	CloudPublicIps      []string      `json:"cloud_public_ips"`
	CloudPublicIps6     []string      `json:"cloud_public_ips6"`
	HostIps             []string      `json:"host_ips"`
	NodePortIps         []string      `json:"node_port_ips"`
	NetworkNamespace    string        `json:"network_namespace"`
	NoPublicAddress     bool          `json:"no_public_address"`
	NoPublicAddress6    bool          `json:"no_public_address6"`
	NoHostAddress       bool          `json:"no_host_address"`
	Node                bson.ObjectID `json:"node"`
	Shape               bson.ObjectID `json:"shape"`
	Name                string        `json:"name"`
	RootEnabled         bool          `json:"root_enabled"`
	Memory              int           `json:"memory"`
	Processors          int           `json:"processors"`
	Roles               []string      `json:"roles"`
	Vnc                 bool          `json:"vnc"`
	Spice               bool          `json:"spice"`
	Gui                 bool          `json:"gui"`
	Deployment          bson.ObjectID `json:"deployment"`
}

func NewInstance(inst *instance.Instance) *Instance {
	if inst == nil {
		return &Instance{}
	}

	return &Instance{
		Id:                  inst.Id,
		Organization:        inst.Organization,
		Zone:                inst.Zone,
		Vpc:                 inst.Vpc,
		Subnet:              inst.Subnet,
		CloudSubnet:         inst.CloudSubnet,
		CloudVnic:           inst.CloudVnic,
		Image:               inst.Image,
		State:               inst.State,
		Timestamp:           inst.Timestamp,
		Action:              inst.Action,
		Uefi:                inst.Uefi,
		SecureBoot:          inst.SecureBoot,
		Tpm:                 inst.Tpm,
		DhcpServer:          inst.DhcpServer,
		CloudType:           inst.CloudType,
		SystemKind:          inst.SystemKind,
		DeleteProtection:    inst.DeleteProtection,
		SkipSourceDestCheck: inst.SkipSourceDestCheck,
		QemuVersion:         inst.QemuVersion,
		PublicIps:           inst.PublicIps,
		PublicIps6:          inst.PublicIps6,
		PrivateIps:          inst.PrivateIps,
		PrivateIps6:         inst.PrivateIps6,
		GatewayIps:          inst.GatewayIps,
		GatewayIps6:         inst.GatewayIps6,
		CloudPrivateIps:     inst.CloudPrivateIps,
		CloudPublicIps:      inst.CloudPublicIps,
		CloudPublicIps6:     inst.CloudPublicIps6,
		HostIps:             inst.HostIps,
		NodePortIps:         inst.NodePortIps,
		NetworkNamespace:    inst.NetworkNamespace,
		NoPublicAddress:     inst.NoPublicAddress,
		NoPublicAddress6:    inst.NoPublicAddress6,
		NoHostAddress:       inst.NoHostAddress,
		Node:                inst.Node,
		Shape:               inst.Shape,
		Name:                inst.Name,
		RootEnabled:         inst.RootEnabled,
		Memory:              inst.Memory,
		Processors:          inst.Processors,
		Roles:               inst.Roles,
		Vnc:                 inst.Vnc,
		Spice:               inst.Spice,
		Gui:                 inst.Gui,
		Deployment:          inst.Deployment,
	}
}

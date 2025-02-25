package types

import (
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/instance"
)

type Instance struct {
	Id                  primitive.ObjectID `json:"id"`
	Organization        primitive.ObjectID `json:"organization"`
	Zone                primitive.ObjectID `json:"zone"`
	Vpc                 primitive.ObjectID `json:"vpc"`
	Subnet              primitive.ObjectID `json:"subnet"`
	OracleSubnet        string             `json:"oracle_subnet"`
	OracleVnic          string             `json:"oracle_vnic"`
	Image               primitive.ObjectID `json:"image"`
	State               string             `json:"state"`
	Uefi                bool               `json:"uefi"`
	SecureBoot          bool               `json:"secure_boot"`
	Tpm                 bool               `json:"tpm"`
	DhcpServer          bool               `json:"dhcp_server"`
	CloudType           string             `json:"cloud_type"`
	DeleteProtection    bool               `json:"delete_protection"`
	SkipSourceDestCheck bool               `json:"skip_source_dest_check"`
	QemuVersion         string             `json:"qemu_version"`
	PublicIps           []string           `json:"public_ips"`
	PublicIps6          []string           `json:"public_ips6"`
	PrivateIps          []string           `json:"private_ips"`
	PrivateIps6         []string           `json:"private_ips6"`
	GatewayIps          []string           `json:"gateway_ips"`
	GatewayIps6         []string           `json:"gateway_ips6"`
	OraclePrivateIps    []string           `json:"oracle_private_ips"`
	OraclePublicIps     []string           `json:"oracle_public_ips"`
	HostIps             []string           `json:"host_ips"`
	NodePortIps         []string           `json:"node_port_ips"`
	NetworkNamespace    string             `json:"network_namespace"`
	NoPublicAddress     bool               `json:"no_public_address"`
	NoPublicAddress6    bool               `json:"no_public_address6"`
	NoHostAddress       bool               `json:"no_host_address"`
	Node                primitive.ObjectID `json:"node"`
	Shape               primitive.ObjectID `json:"shape"`
	Name                string             `json:"name"`
	RootEnabled         bool               `json:"root_enabled"`
	Memory              int                `json:"memory"`
	Processors          int                `json:"processors"`
	NetworkRoles        []string           `json:"network_roles"`
	Vnc                 bool               `json:"vnc"`
	Spice               bool               `json:"spice"`
	Gui                 bool               `json:"gui"`
	Deployment          primitive.ObjectID `json:"deployment"`
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
		OracleSubnet:        inst.OracleSubnet,
		OracleVnic:          inst.OracleVnic,
		Image:               inst.Image,
		State:               inst.State,
		Uefi:                inst.Uefi,
		SecureBoot:          inst.SecureBoot,
		Tpm:                 inst.Tpm,
		DhcpServer:          inst.DhcpServer,
		CloudType:           inst.CloudType,
		DeleteProtection:    inst.DeleteProtection,
		SkipSourceDestCheck: inst.SkipSourceDestCheck,
		QemuVersion:         inst.QemuVersion,
		PublicIps:           inst.PublicIps,
		PublicIps6:          inst.PublicIps6,
		PrivateIps:          inst.PrivateIps,
		PrivateIps6:         inst.PrivateIps6,
		GatewayIps:          inst.GatewayIps,
		GatewayIps6:         inst.GatewayIps6,
		OraclePrivateIps:    inst.OraclePrivateIps,
		OraclePublicIps:     inst.OraclePublicIps,
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
		NetworkRoles:        inst.NetworkRoles,
		Vnc:                 inst.Vnc,
		Spice:               inst.Spice,
		Gui:                 inst.Gui,
		Deployment:          inst.Deployment,
	}
}

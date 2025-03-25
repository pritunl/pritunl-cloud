package spec

import (
	"github.com/pritunl/mongo-go-driver/bson/primitive"
)

type Instance struct {
	Plan                primitive.ObjectID   `bson:"plan,omitempty" json:"plan"`                           // clear
	Datacenter          primitive.ObjectID   `bson:"datacenter" json:"datacenter"`                         // hard
	Zone                primitive.ObjectID   `bson:"zone" json:"zone"`                                     // hard
	Node                primitive.ObjectID   `bson:"node,omitempty" json:"node"`                           // hard
	Shape               primitive.ObjectID   `bson:"shape,omitempty" json:"shape"`                         // hard
	Vpc                 primitive.ObjectID   `bson:"vpc" json:"vpc"`                                       // hard
	Subnet              primitive.ObjectID   `bson:"subnet" json:"subnet"`                                 // hard
	Roles               []string             `bson:"roles" json:"roles"`                                   // soft
	Processors          int                  `bson:"processors" json:"processors"`                         // soft
	Memory              int                  `bson:"memory" json:"memory"`                                 // soft
	Uefi                *bool                `bson:"uefi,omitempty" json:"uefi"`                           // soft
	SecureBoot          *bool                `bson:"secure_boot,omitempty" json:"secure_boot"`             // soft
	CloudType           string               `bson:"cloud_type" json:"cloud_type"`                         // soft
	Tpm                 bool                 `bson:"tpm" json:"tpm"`                                       // soft
	Vnc                 bool                 `bson:"vnc" json:"vnc"`                                       // soft
	DeleteProtection    bool                 `bson:"delete_protection" json:"delete_protection"`           // soft
	SkipSourceDestCheck bool                 `bson:"skip_source_dest_check" json:"skip_source_dest_check"` // soft
	HostAddress         *bool                `bson:"host_address,omitempty" json:"host_address"`           // soft
	PublicAddress       *bool                `bson:"public_address,omitempty" json:"public_address"`       // soft
	PublicAddress6      *bool                `bson:"public_address6,omitempty" json:"public_address6"`     // soft
	DhcpServer          bool                 `bson:"dhcp_server" json:"dhcp_server"`                       // soft
	Image               primitive.ObjectID   `bson:"image" json:"image"`                                   // hard
	DiskSize            int                  `bson:"disk_size" json:"disk_size"`                           // hard
	Mounts              []Mount              `bson:"mounts" json:"mounts"`                                 // hard
	NodePorts           []NodePort           `bson:"node_ports" json:"node_ports"`                         // soft
	Certificates        []primitive.ObjectID `bson:"certificates" json:"certificates"`                     // soft
	Secrets             []primitive.ObjectID `bson:"secrets" json:"secrets"`                               // soft
	Pods                []primitive.ObjectID `bson:"pods" json:"pods"`                                     // soft
}

type NodePort struct {
	Protocol     string `bson:"protocol" json:"protocol"`
	ExternalPort int    `bson:"external_port" json:"external_port"`
	InternalPort int    `bson:"internal_port" json:"internal_port"`
}

func (i *Instance) MemoryUnits() float64 {
	return float64(i.Memory) / float64(1024)
}

type Mount struct {
	Path  string               `bson:"path" json:"path"`
	Disks []primitive.ObjectID `bson:"disks" json:"disks"`
}

type InstanceYaml struct {
	Name                string              `yaml:"name"`
	Kind                string              `yaml:"kind"`
	Count               int                 `yaml:"count"`
	Plan                string              `yaml:"plan"`
	Zone                string              `yaml:"zone"`
	Node                string              `yaml:"node,omitempty"`
	Shape               string              `yaml:"shape,omitempty"`
	Vpc                 string              `yaml:"vpc"`
	Subnet              string              `yaml:"subnet"`
	Roles               []string            `yaml:"roles"`
	Processors          int                 `yaml:"processors"`
	Memory              int                 `yaml:"memory"`
	Uefi                *bool               `yaml:"uefi"`
	SecureBoot          *bool               `yaml:"secure_boot"`
	CloudType           string              `yaml:"cloud_type"`
	Tpm                 bool                `yaml:"tpm"`
	Vnc                 bool                `yaml:"vnc"`
	DeleteProtection    bool                `yaml:"delete_protection"`
	SkipSourceDestCheck bool                `yaml:"skip_source_dest_check"`
	HostAddress         *bool               `yaml:"host_address"`
	PublicAddress       *bool               `yaml:"public_address"`
	PublicAddress6      *bool               `yaml:"public_address6"`
	DhcpServer          bool                `yaml:"dhcp_server"`
	Image               string              `yaml:"image"`
	Mounts              []InstanceMountYaml `yaml:"mounts"`
	Certificates        []string            `yaml:"certificates"`
	Secrets             []string            `yaml:"secrets"`
	Pods                []string            `yaml:"pods"`
	DiskSize            int                 `yaml:"disk-size"`
}

type InstanceMountYaml struct {
	Path  string   `yaml:"path"`
	Disks []string `yaml:"disks"`
}

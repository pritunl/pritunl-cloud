package spec

import (
	"github.com/pritunl/mongo-go-driver/bson/primitive"
)

type Instance struct {
	Plan         primitive.ObjectID   `bson:"plan,omitempty" json:"plan"`       // clear
	Datacenter   primitive.ObjectID   `bson:"datacenter" json:"datacenter"`     // hard
	Zone         primitive.ObjectID   `bson:"zone" json:"zone"`                 // hard
	Node         primitive.ObjectID   `bson:"node,omitempty" json:"node"`       // hard
	Shape        primitive.ObjectID   `bson:"shape,omitempty" json:"shape"`     // hard
	Vpc          primitive.ObjectID   `bson:"vpc" json:"vpc"`                   // hard
	Subnet       primitive.ObjectID   `bson:"subnet" json:"subnet"`             // hard
	Roles        []string             `bson:"roles" json:"roles"`               // soft
	Processors   int                  `bson:"processors" json:"processors"`     // soft
	Memory       int                  `bson:"memory" json:"memory"`             // soft
	Image        primitive.ObjectID   `bson:"image" json:"image"`               // hard
	DiskSize     int                  `bson:"disk_size" json:"disk_size"`       // hard
	Mounts       []Mount              `bson:"mounts" json:"mounts"`             // hard
	Certificates []primitive.ObjectID `bson:"certificates" json:"certificates"` // soft
	Secrets      []primitive.ObjectID `bson:"secrets" json:"secrets"`           // soft
	Pods         []primitive.ObjectID `bson:"pods" json:"pods"`                 // soft
}

func (i *Instance) MemoryUnits() float64 {
	return float64(i.Memory) / float64(1024)
}

type Mount struct {
	Path  string               `bson:"path" json:"path"`
	Disks []primitive.ObjectID `bson:"disks" json:"disks"`
}

type InstanceYaml struct {
	Name         string              `yaml:"name"`
	Kind         string              `yaml:"kind"`
	Count        int                 `yaml:"count"`
	Plan         string              `yaml:"plan"`
	Zone         string              `yaml:"zone"`
	Node         string              `yaml:"node,omitempty"`
	Shape        string              `yaml:"shape,omitempty"`
	Vpc          string              `yaml:"vpc"`
	Subnet       string              `yaml:"subnet"`
	Roles        []string            `yaml:"roles"`
	Processors   int                 `yaml:"processors"`
	Memory       int                 `yaml:"memory"`
	Image        string              `yaml:"image"`
	Mounts       []InstanceMountYaml `yaml:"mounts"`
	Certificates []string            `yaml:"certificates"`
	Secrets      []string            `yaml:"secrets"`
	Pods         []string            `yaml:"pods"`
	DiskSize     int                 `yaml:"disk-size"`
}

type InstanceMountYaml struct {
	Path  string   `yaml:"path"`
	Disks []string `yaml:"disks"`
}

package service

import (
	"github.com/pritunl/mongo-go-driver/bson/primitive"
)

type Unit struct {
	Name string      `bson:"name" json:"name"`
	Kind string      `bson:"kind" json:"kind"`
	Data interface{} `bson:"data" json:"data"`
	Spec string      `bson:"spec" json:"spec"`
}

type Instance struct {
	Zone       primitive.ObjectID `bson:"zone" json:"zone"`
	Node       primitive.ObjectID `bson:"node,omitempty" json:"node"`
	Shape      primitive.ObjectID `bson:"shape,omitempty" json:"shape"`
	Vpc        primitive.ObjectID `bson:"vpc" json:"vpc"`
	Subnet     primitive.ObjectID `bson:"subnet" json:"subnet"`
	Roles      []string           `bson:"roles" json:"roles"`
	Processors int                `bson:"processors" json:"processors"`
	Memory     int                `bson:"memory" json:"memory"`
	Image      primitive.ObjectID `bson:"image" json:"image"`
	DiskSize   int                `bson:"disk_size" json:"disk_size"`
}

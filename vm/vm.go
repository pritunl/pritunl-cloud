package vm

import (
	"gopkg.in/mgo.v2/bson"
)

type VirtualMachine struct {
	Id              bson.ObjectId
	Uuid            string
	State           string
	Processors      int
	Memory          int
	Disks           []*Disk
	NetworkAdapters []*NetworkAdapter
}

type Disk struct {
	Path string
}

type NetworkAdapter struct {
	MacAddress       string
	BridgedInterface string
	IpAddress        string
	IpAddress6       string
}

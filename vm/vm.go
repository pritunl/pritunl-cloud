package vm

import (
	"path"
	"strings"
	"time"

	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
)

type VirtualMachine struct {
	Id              primitive.ObjectID `json:"id"`
	State           string             `json:"state"`
	Timestamp       time.Time          `json:"timestamp"`
	Image           primitive.ObjectID `json:"image"`
	Processors      int                `json:"processors"`
	Memory          int                `json:"memory"`
	Vnc             bool               `json:"vnc"`
	VncDisplay      int                `json:"vnc_display"`
	Disks           []*Disk            `json:"disks"`
	NetworkAdapters []*NetworkAdapter  `json:"network_adapters"`
	NoPublicAddress bool               `json:"no_public_address"`
	NoHostAddress   bool               `json:"no_host_address"`
	UsbDevices      []*UsbDevice       `json:"usb_devices"`
}

type Disk struct {
	Index int    `json:"index"`
	Path  string `json:"path"`
}

type UsbDevice struct {
	Vendor  string `json:"vendor"`
	Product string `json:"product"`
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

	err = coll.UpdateId(v.Id, &bson.M{
		"$set": &bson.M{
			"vm_state":     v.State,
			"vm_timestamp": v.Timestamp,
			"public_ips":   addrs,
			"public_ips6":  addrs6,
		},
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

func (v *VirtualMachine) CommitState(db *database.Database, state string) (
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

	err = coll.UpdateId(v.Id, &bson.M{
		"$set": &bson.M{
			"state":        state,
			"vm_state":     v.State,
			"vm_timestamp": v.Timestamp,
			"public_ips":   addrs,
			"public_ips6":  addrs6,
		},
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

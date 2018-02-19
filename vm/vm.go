package vm

import (
	"github.com/pritunl/pritunl-cloud/database"
	"gopkg.in/mgo.v2/bson"
)

type VirtualMachine struct {
	Id              bson.ObjectId     `json:"id"`
	Uuid            string            `json:"uuid"`
	State           string            `json:"state"`
	Image           bson.ObjectId     `json:"image"`
	Processors      int               `json:"processors"`
	Memory          int               `json:"memory"`
	Disks           []*Disk           `json:"disks"`
	NetworkAdapters []*NetworkAdapter `json:"network_adapters"`
}

type Disk struct {
	Path string `json:"path"`
}

type NetworkAdapter struct {
	MacAddress       string `json:"mac_address"`
	BridgedInterface string `json:"bridged_interface"`
	IpAddress        string `json:"ip_address"`
	IpAddress6       string `json:"ip_address6"`
}

func (v *VirtualMachine) Commit(db *database.Database) (err error) {
	coll := db.Instances()

	addr := ""
	addr6 := ""
	if len(v.NetworkAdapters) > 0 {
		addr = v.NetworkAdapters[0].IpAddress
		addr6 = v.NetworkAdapters[0].IpAddress6
	}

	err = coll.UpdateId(v.Id, &bson.M{
		"$set": &bson.M{
			"vm_state":   v.State,
			"public_ip":  addr,
			"public_ip6": addr6,
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

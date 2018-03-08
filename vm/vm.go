package vm

import (
	"github.com/pritunl/pritunl-cloud/database"
	"gopkg.in/mgo.v2/bson"
	"path"
	"strings"
)

type VirtualMachine struct {
	Id              bson.ObjectId     `json:"id"`
	State           string            `json:"state"`
	Image           bson.ObjectId     `json:"image"`
	Processors      int               `json:"processors"`
	Memory          int               `json:"memory"`
	Disks           []*Disk           `json:"disks"`
	NetworkAdapters []*NetworkAdapter `json:"network_adapters"`
}

type Disk struct {
	Index int    `json:"index"`
	Path  string `json:"path"`
}

func (d *Disk) GetId() bson.ObjectId {
	idStr := strings.Split(path.Base(d.Path), ".")[0]
	if bson.IsObjectIdHex(idStr) {
		return bson.ObjectIdHex(idStr)
	}
	return ""
}

type NetworkAdapter struct {
	Type          string        `json:"type"`
	MacAddress    string        `json:"mac_address"`
	HostInterface string        `json:"host_interface"`
	VpcId         bson.ObjectId `json:"vpc_id"`
	IpAddress     string        `json:"ip_address,omitempty"`
	IpAddress6    string        `json:"ip_address6,omitempty"`
}

func (v *VirtualMachine) Commit(db *database.Database) (err error) {
	coll := db.Instances()

	addr := ""
	addr6 := ""
	localAddrs := []string{}
	localAddrs6 := []string{}

	if len(v.NetworkAdapters) > 0 {
		addr = v.NetworkAdapters[0].IpAddress
		addr6 = v.NetworkAdapters[0].IpAddress6

		for _, adapter := range v.NetworkAdapters[1:] {
			localAddrs = append(localAddrs, adapter.IpAddress)
			localAddrs6 = append(localAddrs, adapter.IpAddress6)
		}
	}

	err = coll.UpdateId(v.Id, &bson.M{
		"$set": &bson.M{
			"vm_state":   v.State,
			"public_ip":  addr,
			"public_ip6": addr6,
			"local_ips":  localAddrs,
			"local_ips6": localAddrs6,
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

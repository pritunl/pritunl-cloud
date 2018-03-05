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
	Type          string `json:"type"`
	MacAddress    string `json:"mac_address"`
	HostInterface string `json:"host_interface"`
	IpAddress     string `json:"ip_address,omitempty"`
	IpAddress6    string `json:"ip_address6,omitempty"`
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

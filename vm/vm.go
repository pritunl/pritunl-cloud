package vm

import (
	"github.com/pritunl/pritunl-cloud/database"
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
			"state":      v.State,
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

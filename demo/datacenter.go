package demo

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/datacenter"
	"github.com/pritunl/pritunl-cloud/utils"
)

var Datacenters = []*datacenter.Datacenter{
	{
		Id:                 utils.ObjectIdHex("689733b7a7a35eae0dbaea1b"),
		Name:               "us-west-1",
		Comment:            "",
		MatchOrganizations: false,
		Organizations:      []bson.ObjectID{},
		NetworkMode:        "vxlan_vlan",
		WgMode:             "",
		PublicStorages: []bson.ObjectID{
			utils.ObjectIdHex("689733b7a7a35eae0dbaea15"),
		},
		PrivateStorage:      bson.ObjectID{},
		PrivateStorageClass: "",
		BackupStorage:       bson.ObjectID{},
		BackupStorageClass:  "",
	},
}

package demo

import (
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/datacenter"
	"github.com/pritunl/pritunl-cloud/utils"
)

var Datacenters = []*datacenter.Datacenter{
	{
		Id:                 utils.ObjectIdHex("689733b7a7a35eae0dbaea1b"),
		Name:               "us-west-1",
		Comment:            "",
		MatchOrganizations: false,
		Organizations:      []primitive.ObjectID{},
		NetworkMode:        "vxlan_vlan",
		WgMode:             "",
		PublicStorages: []primitive.ObjectID{
			utils.ObjectIdHex("689733b7a7a35eae0dbaea15"),
		},
		PrivateStorage:      primitive.ObjectID{},
		PrivateStorageClass: "",
		BackupStorage:       primitive.ObjectID{},
		BackupStorageClass:  "",
	},
}

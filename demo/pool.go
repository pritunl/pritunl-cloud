package demo

import (
	"github.com/pritunl/pritunl-cloud/pool"
	"github.com/pritunl/pritunl-cloud/utils"
)

var Pools = []*pool.Pool{
	{
		Id:               utils.ObjectIdHex("67b89e8d4866ba90e6c459ba"),
		Name:             "cloud-east",
		Comment:          "",
		DeleteProtection: false,
		Datacenter:       utils.ObjectIdHex("689733b7a7a35eae0dbaea1b"),
		Zone:             utils.ObjectIdHex("689733b7a7a35eae0dbaea1e"),
		Type:             "",
		VgName:           "cloud_east",
	},
}

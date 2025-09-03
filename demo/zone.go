package demo

import (
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/zone"
)

var Zones = []*zone.Zone{
	{
		Id:          utils.ObjectIdHex("689733b7a7a35eae0dbaea1e"),
		Datacenter:  utils.ObjectIdHex("689733b7a7a35eae0dbaea1b"),
		Name:        "us-west-1a",
		Comment:     "",
		DnsServers:  []string{},
		DnsServers6: []string{},
	},
}

package demo

import (
	"github.com/pritunl/pritunl-cloud/aggregate"
	"github.com/pritunl/pritunl-cloud/block"
	"github.com/pritunl/pritunl-cloud/utils"
)

var Blocks = []*aggregate.BlockAggregate{
	{
		Block: block.Block{
			Id:      utils.ObjectIdHex("689733b7a7a35eae0dbaea2f"),
			Name:    "east-public",
			Comment: "",
			Type:    "ipv4",
			Vlan:    0,
			Subnets: []string{
				"1.253.67.0/24",
			},
			Subnets6: []string{},
			Excludes: []string{
				"1.253.67.90/24",
				"1.253.67.91/24",
				"1.253.67.92/24",
				"1.253.67.93/24",
				"1.253.67.94/24",
				"1.253.67.95/24",
			},
			Netmask:  "255.255.255.0",
			Gateway:  "1.253.67.1",
			Gateway6: "",
		},
		Available: 248,
		Capacity:  254,
	},
	{
		Block: block.Block{
			Id:      utils.ObjectIdHex("68973a47b5844593cf99cc7a"),
			Name:    "east-public6",
			Comment: "",
			Type:    "ipv6",
			Vlan:    0,
			Subnets: []string{},
			Subnets6: []string{
				"2001:db8:85a3:4d2f::/64",
			},
			Excludes: []string{},
			Netmask:  "",
			Gateway:  "",
			Gateway6: "2001:db8:85a3:4d2f::1",
		},
	},
}

package demo

import (
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vpc"
)

var Vpcs = []*vpc.Vpc{
	{
		Id:       utils.ObjectIdHex("689733b7a7a35eae0dbaea23"),
		Name:     "production",
		Comment:  "",
		VpcId:    2996,
		Network:  "10.196.0.0/14",
		Network6: "fd97:30bf:d456:a3bc::/64",
		Subnets: []*vpc.Subnet{
			{
				Id:      utils.ObjectIdHex("66a076d5fafc270786e93461"),
				Name:    "primary",
				Network: "10.196.1.0/24",
			},
			{
				Id:      utils.ObjectIdHex("66a076d5fafc270786e93462"),
				Name:    "management",
				Network: "10.196.2.0/24",
			},
			{
				Id:      utils.ObjectIdHex("66a076d5fafc270786e93463"),
				Name:    "link",
				Network: "10.196.3.0/24",
			},
			{
				Id:      utils.ObjectIdHex("66a076d5fafc270786e93464"),
				Name:    "database",
				Network: "10.196.4.0/24",
			},
			{
				Id:      utils.ObjectIdHex("66a076d5fafc270786e93465"),
				Name:    "web",
				Network: "10.196.5.0/24",
			},
			{
				Id:      utils.ObjectIdHex("66a076d5fafc270786e93466"),
				Name:    "search",
				Network: "10.196.6.0/24",
			},
			{
				Id:      utils.ObjectIdHex("66a076d5fafc270786e93467"),
				Name:    "vpn",
				Network: "10.196.7.0/24",
			},
			{
				Id:      utils.ObjectIdHex("66a076d5fafc270786e93468"),
				Name:    "balancer",
				Network: "10.196.8.0/24",
			},
		},
		Organization:  utils.ObjectIdHex("5a3245a50accad1a8a53bc82"),
		Datacenter:    utils.ObjectIdHex("689733b7a7a35eae0dbaea1b"),
		IcmpRedirects: false,
		Routes: []*vpc.Route{
			&vpc.Route{
				Destination: "10.24.0.0/16",
				Target:      "10.196.7.2",
			},
		},
		Maps:             []*vpc.Map{},
		Arps:             []*vpc.Arp{},
		DeleteProtection: false,
	},
	{
		Id:       utils.ObjectIdHex("689733b7a7a35eae0dbaea24"),
		Name:     "testing",
		Comment:  "",
		VpcId:    2732,
		Network:  "10.224.0.0/14",
		Network6: "fd97:30bf:d456:a3bc::/64",
		Subnets: []*vpc.Subnet{
			{
				Id:      utils.ObjectIdHex("689733b7a7a35eae0dbaea61"),
				Name:    "primary",
				Network: "10.224.1.0/24",
			},
			{
				Id:      utils.ObjectIdHex("689733b7a7a35eae0dbaea62"),
				Name:    "management",
				Network: "10.224.2.0/24",
			},
			{
				Id:      utils.ObjectIdHex("689733b7a7a35eae0dbaea63"),
				Name:    "link",
				Network: "10.224.3.0/24",
			},
			{
				Id:      utils.ObjectIdHex("689733b7a7a35eae0dbaea64"),
				Name:    "database",
				Network: "10.224.4.0/24",
			},
			{
				Id:      utils.ObjectIdHex("689733b7a7a35eae0dbaea65"),
				Name:    "web",
				Network: "10.224.5.0/24",
			},
			{
				Id:      utils.ObjectIdHex("689733b7a7a35eae0dbaea66"),
				Name:    "search",
				Network: "10.224.6.0/24",
			},
			{
				Id:      utils.ObjectIdHex("689733b7a7a35eae0dbaea67"),
				Name:    "vpn",
				Network: "10.224.7.0/24",
			},
			{
				Id:      utils.ObjectIdHex("689733b7a7a35eae0dbaea68"),
				Name:    "balancer",
				Network: "10.224.8.0/24",
			},
		},
		Organization:  utils.ObjectIdHex("5a3245a50accad1a8a53bc82"),
		Datacenter:    utils.ObjectIdHex("689733b7a7a35eae0dbaea1b"),
		IcmpRedirects: false,
		Routes: []*vpc.Route{
			&vpc.Route{
				Destination: "10.36.0.0/16",
				Target:      "10.224.7.2",
			},
		},
		Maps:             []*vpc.Map{},
		Arps:             []*vpc.Arp{},
		DeleteProtection: false,
	},
}

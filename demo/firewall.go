package demo

import (
	"github.com/pritunl/pritunl-cloud/firewall"
	"github.com/pritunl/pritunl-cloud/utils"
)

var Firewalls = []*firewall.Firewall{
	{
		Id:           utils.ObjectIdHex("688ab80d1793930f821f4f39"),
		Name:         "instance",
		Comment:      "",
		Organization: utils.ObjectIdHex("5a3245a50accad1a8a53bc82"),
		Roles:        []string{"instance"},
		Ingress: []*firewall.Rule{
			{
				SourceIps: []string{"0.0.0.0/0", "::/0"},
				Protocol:  "icmp",
				Port:      "",
			},
			{
				SourceIps: []string{"0.0.0.0/0", "::/0"},
				Protocol:  "tcp",
				Port:      "22",
			},
			{
				SourceIps: []string{"0.0.0.0/0", "::/0"},
				Protocol:  "tcp",
				Port:      "80",
			},
			{
				SourceIps: []string{"0.0.0.0/0", "::/0"},
				Protocol:  "tcp",
				Port:      "443",
			},
		},
	},
}

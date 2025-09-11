package demo

import (
	"time"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/balancer"
	"github.com/pritunl/pritunl-cloud/utils"
)

var Balancers = []*balancer.Balancer{
	{
		Id:           utils.ObjectIdHex("61ba27ccf149d4c222b23247"),
		Name:         "web-app",
		Comment:      "",
		Type:         "http",
		State:        true,
		Organization: utils.ObjectIdHex("5a3245a50accad1a8a53bc82"),
		Datacenter:   utils.ObjectIdHex("689733b7a7a35eae0dbaea1b"),
		Certificates: []bson.ObjectID{
			utils.ObjectIdHex("67b89ef24866ba90e6c459e8"),
		},
		ClientAuthority: bson.ObjectID{},
		WebSockets:      false,
		Domains: []*balancer.Domain{
			{
				Domain: "demo.cloud.pritunl.com",
				Host:   "",
			},
		},
		Backends: []*balancer.Backend{
			{
				Protocol: "http",
				Hostname: "10.234.10.22",
				Port:     8000,
			},
			{
				Protocol: "http",
				Hostname: "10.234.10.24",
				Port:     8000,
			},
		},
		States: map[string]*balancer.State{
			"65b5d7e1c2e9a21159765955": {
				Timestamp:  time.Now(),
				Requests:   125,
				Retries:    0,
				WebSockets: 0,
				Online: []string{
					"10.234.10.22:8000",
					"10.234.10.24:8000",
				},
				UnknownHigh: []string{},
				UnknownMid:  []string{},
				UnknownLow:  []string{},
				Offline:     []string{},
			},
		},
		CheckPath: "/check",
	},
}

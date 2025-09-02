package demo

import (
	"github.com/pritunl/pritunl-cloud/plan"
	"github.com/pritunl/pritunl-cloud/utils"
)

var Plans = []*plan.Plan{
	{
		Id:           utils.ObjectIdHex("66e8993f1fbc6db8e20819f8"),
		Name:         "primary",
		Comment:      "",
		Organization: utils.ObjectIdHex("5a3245a50accad1a8a53bc82"),
		Statements: []*plan.Statement{
			{
				Id:        utils.ObjectIdHex("67c9bed42c125c5ddf24d0a1"),
				Statement: "IF instance.last_timestamp < 60 AND instance.last_heartbeat > 60 FOR 15 THEN 'stop'",
			},
			{
				Id:        utils.ObjectIdHex("683d645e2956cdd93d3e08d2"),
				Statement: "IF instance.state != 'running' THEN 'start'",
			},
		},
	},
}

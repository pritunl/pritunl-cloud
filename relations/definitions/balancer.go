package definitions

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/relations"
)

var Balancer = relations.Query{
	Label:      "Load Balancer",
	Collection: "balancers",
	Project: []relations.Project{{
		Key:   "name",
		Label: "Name",
	}, {
		Key:   "type",
		Label: "Type",
	}, {
		Key:   "state",
		Label: "State",
	}, {
		Key: "datacenter",
	}},
	Relations: []relations.Relation{{
		Key:          "nodes",
		Label:        "Node",
		From:         "nodes",
		LocalField:   "datacenter",
		ForeignField: "datacenter",
		Match: bson.M{
			"types": node.Balancer,
		},
		Sort: map[string]int{
			"name": 1,
		},
		Project: []relations.Project{{
			Key:   "name",
			Label: "Name",
		}, {
			Key:   "types",
			Label: "Modes",
		}, {
			Key:   "admin_domain",
			Label: "Admin Domain",
		}, {
			Key:   "user_domain",
			Label: "User Domain",
		}, {
			Key:   "network_mode",
			Label: "Network Mode IPv4",
		}, {
			Key:   "network_mode6",
			Label: "Network Mode IPv6",
		}},
	}},
}

func init() {
	relations.Register("balancer", Balancer)
}

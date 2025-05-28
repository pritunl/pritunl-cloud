package definitions

import (
	"github.com/pritunl/pritunl-cloud/relations"
)

var Certificate = relations.Query{
	Label:      "Certificate",
	Collection: "certificates",
	Project: []relations.Project{{
		Key:   "name",
		Label: "Name",
	}, {
		Key:   "roles",
		Label: "Roles",
	}},
	Relations: []relations.Relation{{
		Key:          "nodes",
		Label:        "Node",
		From:         "nodes",
		LocalField:   "_id",
		ForeignField: "certificates",
		BlockDelete:  true,
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
			Key:   "webauthn_domain",
			Label: "WebAuthn Domain",
		}, {
			Key:   "network_mode",
			Label: "Network Mode IPv4",
		}, {
			Key:   "network_mode6",
			Label: "Network Mode IPv6",
		}},
	}, {
		Key:          "balancers",
		Label:        "Load Balancer",
		From:         "balancers",
		LocalField:   "_id",
		ForeignField: "certificates",
		BlockDelete:  true,
		Sort: map[string]int{
			"name": 1,
		},
		Project: []relations.Project{{
			Key:   "name",
			Label: "Name",
		}, {
			Key:   "state",
			Label: "State",
		}},
	}},
}

func init() {
	relations.Register("certificate", Certificate)
}

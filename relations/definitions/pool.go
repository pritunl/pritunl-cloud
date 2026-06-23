package definitions

import (
	"github.com/pritunl/pritunl-cloud/relations"
)

var Pool = relations.Query{
	Label:      "Pool",
	Collection: "pools",
	Project: []relations.Project{{
		Key:   "name",
		Label: "Name",
	}, {
		Key:   "type",
		Label: "Type",
	}, {
		Key:   "vg_name",
		Label: "VG Name",
	}},
	Relations: []relations.Relation{{
		Key:          "nodes",
		Label:        "Node",
		From:         "nodes",
		LocalField:   "vg_name",
		ForeignField: "pools",
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
			Key:   "network_mode",
			Label: "Network Mode IPv4",
		}, {
			Key:   "network_mode6",
			Label: "Network Mode IPv6",
		}},
	}},
}

func init() {
	relations.Register("pool", Pool)
}

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
		LocalField:   "_id",
		ForeignField: "pools",
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
	}, {
		Key:          "shapes",
		Label:        "Shape",
		From:         "shapes",
		LocalField:   "_id",
		ForeignField: "disk_pools",
		BlockDelete:  true,
		Sort: map[string]int{
			"name": 1,
		},
		Project: []relations.Project{{
			Key:   "name",
			Label: "Name",
		}, {
			Key:   "type",
			Label: "Type",
		}, {
			Key:   "roles",
			Label: "Roles",
		}},
	}, {
		Key:          "disks",
		Label:        "Disk",
		From:         "disks",
		LocalField:   "_id",
		ForeignField: "pool",
		BlockDelete:  true,
		Sort: map[string]int{
			"name": 1,
		},
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
			Key:   "size",
			Label: "Size",
		}},
	}},
}

func init() {
	relations.Register("pool", Pool)
}

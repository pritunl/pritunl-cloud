package definitions

import (
	"github.com/pritunl/pritunl-cloud/relations"
)

var Pod = relations.Query{
	Label:      "Pod",
	Collection: "pods",
	Project: []relations.Project{{
		Key:   "name",
		Label: "Name",
	}},
	Relations: []relations.Relation{{
		Key:          "units",
		Label:        "Unit",
		From:         "units",
		LocalField:   "_id",
		ForeignField: "pod",
		Sort: map[string]int{
			"name": 1,
		},
		Project: []relations.Project{{
			Key:   "name",
			Label: "Name",
		}, {
			Key:   "kind",
			Label: "Kind",
		}, {
			Key:   "count",
			Label: "Count",
		}},
		Relations: []relations.Relation{{
			Key:          "deployments",
			Label:        "Deployment",
			From:         "deployments",
			LocalField:   "_id",
			ForeignField: "unit",
			Sort: map[string]int{
				"name": 1,
			},
			Project: []relations.Project{{
				Key:   "name",
				Label: "Name",
			}, {
				Key:   "kind",
				Label: "Kind",
			}, {
				Key:   "state",
				Label: "State",
			}, {
				Key:   "status",
				Label: "Status",
			}, {
				Key:   "timestamp",
				Label: "Timestamp",
			}},
			Relations: []relations.Relation{{
				Key:          "instances",
				Label:        "Instance",
				From:         "instances",
				LocalField:   "_id",
				ForeignField: "deployment",
				Sort: map[string]int{
					"name": 1,
				},
				Project: []relations.Project{{
					Key:   "name",
					Label: "Name",
				}, {
					Key:   "state",
					Label: "State",
				}, {
					Key:   "state",
					Label: "State",
				}, {
					Key:   "status",
					Label: "Status",
				}, {
					Key:   "virt_state",
					Label: "Virtual State",
				}, {
					Key:   "virt_timestamp",
					Label: "Virtual Timestamp",
				}, {
					Key:   "private_ips",
					Label: "Private IPv4",
				}, {
					Key:   "public_ips",
					Label: "Public IPv4",
				}},
			}, {
				Key:          "disks",
				Label:        "Disk",
				From:         "disks",
				LocalField:   "_id",
				ForeignField: "deployment",
				Sort: map[string]int{
					"index": 1,
				},
				Project: []relations.Project{{
					Key:   "name",
					Label: "Name",
				}, {
					Key:   "index",
					Label: "Index",
				}, {
					Key:   "size",
					Label: "Size",
				}},
			}},
		}},
	}},
}

func init() {
	relations.Register("pod", Pod)
}

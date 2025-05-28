package definitions

import (
	"github.com/pritunl/pritunl-cloud/relations"
)

var Datacenter = relations.Query{
	Label:      "Datacenter",
	Collection: "datacenters",
	Project: []relations.Project{{
		Key:   "name",
		Label: "Name",
	}, {
		Key:   "roles",
		Label: "Roles",
	}},
	Relations: []relations.Relation{{
		Key:          "zones",
		Label:        "Zone",
		From:         "zones",
		LocalField:   "_id",
		ForeignField: "datacenter",
		BlockDelete:  true,
		Sort: map[string]int{
			"name": 1,
		},
		Project: []relations.Project{{
			Key:   "name",
			Label: "Name",
		}},
	}, {
		Key:          "nodes",
		Label:        "Node",
		From:         "nodes",
		LocalField:   "_id",
		ForeignField: "datacenter",
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
		Key:          "vpcs",
		Label:        "VPC",
		From:         "vpcs",
		LocalField:   "_id",
		ForeignField: "datacenter",
		BlockDelete:  true,
		Sort: map[string]int{
			"name": 1,
		},
		Project: []relations.Project{{
			Key:   "name",
			Label: "Name",
		}, {
			Key:   "vpc_id",
			Label: "VPC ID",
		}, {
			Key:   "network",
			Label: "Network IPv4",
		}},
	}, {
		Key:          "balancers",
		Label:        "Load Balancer",
		From:         "balancers",
		LocalField:   "_id",
		ForeignField: "datacenter",
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
	}, {
		Key:          "deployments",
		Label:        "Deployment",
		From:         "deployments",
		LocalField:   "_id",
		ForeignField: "unit",
		BlockDelete:  true,
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
	}, {
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
	}, {
		Key:          "nodeports",
		Label:        "Nodeport",
		From:         "nodeports",
		LocalField:   "_id",
		ForeignField: "datacenter",
		BlockDelete:  true,
		Sort: map[string]int{
			"port": 1,
		},
		Project: []relations.Project{{
			Key:   "port",
			Label: "Port",
		}, {
			Key:   "protocol",
			Label: "Protocol",
		}},
	}},
}

func init() {
	relations.Register("datacenter", Datacenter)
}

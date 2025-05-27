package definitions

import (
	"github.com/pritunl/pritunl-cloud/relations"
)

var Organization = relations.Query{
	Label:      "Organization",
	Collection: "organizations",
	Project: []relations.Project{{
		Key:   "name",
		Label: "Name",
	}, {
		Key:   "roles",
		Label: "Roles",
	}},
	Relations: []relations.Relation{{
		Key:          "certificates",
		Label:        "Certificate",
		From:         "certificates",
		LocalField:   "_id",
		ForeignField: "organization",
		BlockDelete:  true,
		Sort: map[string]int{
			"name": 1,
		},
		Project: []relations.Project{{
			Key:   "name",
			Label: "Name",
		}},
	}, {
		Key:          "secrets",
		Label:        "Secret",
		From:         "secrets",
		LocalField:   "_id",
		ForeignField: "organization",
		BlockDelete:  true,
		Sort: map[string]int{
			"name": 1,
		},
		Project: []relations.Project{{
			Key:   "name",
			Label: "Name",
		}},
	}, {
		Key:          "vpc",
		Label:        "VPCs",
		From:         "vpc",
		LocalField:   "_id",
		ForeignField: "organization",
		BlockDelete:  true,
		Sort: map[string]int{
			"name": 1,
		},
		Project: []relations.Project{{
			Key:   "name",
			Label: "Name",
		}},
	}, {
		Key:          "domains",
		Label:        "Domain",
		From:         "domains",
		LocalField:   "_id",
		ForeignField: "organization",
		BlockDelete:  true,
		Sort: map[string]int{
			"name": 1,
		},
		Project: []relations.Project{{
			Key:   "name",
			Label: "Name",
		}},
	}, {
		Key:          "balancers",
		Label:        "Load Balancer",
		From:         "balancers",
		LocalField:   "_id",
		ForeignField: "organization",
		BlockDelete:  true,
		Sort: map[string]int{
			"name": 1,
		},
		Project: []relations.Project{{
			Key:   "name",
			Label: "Name",
		}},
	}, {
		Key:          "images",
		Label:        "Image",
		From:         "images",
		LocalField:   "_id",
		ForeignField: "organization",
		BlockDelete:  true,
		Sort: map[string]int{
			"name": 1,
		},
		Project: []relations.Project{{
			Key:   "name",
			Label: "Name",
		}},
	}, {
		Key:          "plans",
		Label:        "Plan",
		From:         "plans",
		LocalField:   "_id",
		ForeignField: "organization",
		BlockDelete:  true,
		Sort: map[string]int{
			"name": 1,
		},
		Project: []relations.Project{{
			Key:   "name",
			Label: "Name",
		}},
	}, {
		Key:          "disks",
		Label:        "Disk",
		From:         "disks",
		LocalField:   "_id",
		ForeignField: "organization",
		BlockDelete:  true,
		Sort: map[string]int{
			"name": 1,
		},
		Project: []relations.Project{{
			Key:   "name",
			Label: "Name",
		}},
	}, {
		Key:          "instances",
		Label:        "Instance",
		From:         "instances",
		LocalField:   "_id",
		ForeignField: "organization",
		BlockDelete:  true,
		Sort: map[string]int{
			"name": 1,
		},
		Project: []relations.Project{{
			Key:   "name",
			Label: "Name",
		}},
	}, {
		Key:          "pods",
		Label:        "Pod",
		From:         "pods",
		LocalField:   "_id",
		ForeignField: "organization",
		BlockDelete:  true,
		Sort: map[string]int{
			"name": 1,
		},
		Project: []relations.Project{{
			Key:   "name",
			Label: "Name",
		}},
	}, {
		Key:          "deployments",
		Label:        "Deployment",
		From:         "deployments",
		LocalField:   "_id",
		ForeignField: "organization",
		BlockDelete:  true,
		Sort: map[string]int{
			"name": 1,
		},
		Project: []relations.Project{{
			Key:   "name",
			Label: "Name",
		}},
	}, {
		Key:          "firewalls",
		Label:        "Firewall",
		From:         "firewalls",
		LocalField:   "_id",
		ForeignField: "organization",
		BlockDelete:  true,
		Sort: map[string]int{
			"name": 1,
		},
		Project: []relations.Project{{
			Key:   "name",
			Label: "Name",
		}},
	}, {
		Key:          "authorities",
		Label:        "Authority",
		From:         "authorities",
		LocalField:   "_id",
		ForeignField: "organization",
		BlockDelete:  true,
		Sort: map[string]int{
			"name": 1,
		},
		Project: []relations.Project{{
			Key:   "name",
			Label: "Name",
		}},
	}, {
		Key:          "alerts",
		Label:        "Alert",
		From:         "alerts",
		LocalField:   "_id",
		ForeignField: "organization",
		BlockDelete:  true,
		Sort: map[string]int{
			"name": 1,
		},
		Project: []relations.Project{{
			Key:   "name",
			Label: "Name",
		}},
	}},
}

func init() {
	relations.Register("organization", Organization)
}

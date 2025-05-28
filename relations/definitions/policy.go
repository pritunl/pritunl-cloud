package definitions

import (
	"github.com/pritunl/pritunl-cloud/relations"
)

var Policy = relations.Query{
	Label:      "Policy",
	Collection: "policies",
	Project: []relations.Project{{
		Key:   "name",
		Label: "Name",
	}, {
		Key:   "roles",
		Label: "Roles",
	}},
	Relations: []relations.Relation{{
		Key:          "users",
		Label:        "User",
		From:         "users",
		LocalField:   "roles",
		ForeignField: "roles",
		Sort: map[string]int{
			"username": 1,
		},
		Project: []relations.Project{{
			Key:   "username",
			Label: "Username",
		}, {
			Key:   "type",
			Label: "Type",
		}},
	}},
}

func init() {
	relations.Register("policy", Policy)
}

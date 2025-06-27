package definitions

import (
	"github.com/pritunl/pritunl-cloud/relations"
)

var Secret = relations.Query{
	Label:      "Secret",
	Collection: "secrets",
	Project: []relations.Project{{
		Key:   "name",
		Label: "Name",
	}},
	Relations: []relations.Relation{{
		Key:          "domains",
		Label:        "Domain",
		From:         "domains",
		LocalField:   "_id",
		ForeignField: "secret",
		BlockDelete:  true,
		Sort: map[string]int{
			"name": 1,
		},
		Project: []relations.Project{{
			Key:   "name",
			Label: "Name",
		}, {
			Key:   "root_domain",
			Label: "Root Domain",
		}},
	}, {
		Key:          "certificates",
		Label:        "Certificate",
		From:         "certificates",
		LocalField:   "_id",
		ForeignField: "acme_secret",
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
			Key:   "acme_domains",
			Label: "Lets Encrypt Domains",
		}},
	}},
}

func init() {
	relations.Register("secret", Secret)
}

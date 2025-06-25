package definitions

import (
	"github.com/pritunl/pritunl-cloud/relations"
	"github.com/pritunl/pritunl-cloud/utils"
)

var Block = relations.Query{
	Label:      "Block",
	Collection: "blocks",
	Project: []relations.Project{{
		Key:   "name",
		Label: "Name",
	}, {
		Key:   "subnets",
		Label: "Subnets",
	}},
	Relations: []relations.Relation{{
		Key:          "blocks_ip",
		Label:        "Block IP",
		From:         "blocks_ip",
		LocalField:   "_id",
		ForeignField: "block",
		Sort: map[string]int{
			"ip": 1,
		},
		Project: []relations.Project{{
			Key:   "name",
			Label: "Name",
		}, {
			Key:   "ip",
			Label: "IP",
			Format: func(vals ...any) any {
				return utils.Int2IpAddress(vals[0].(int64)).String()
			},
		}},
	}},
}

func init() {
	relations.Register("block", Block)
}

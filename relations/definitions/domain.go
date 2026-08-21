package definitions

import (
	"time"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/relations"
	"github.com/pritunl/pritunl-cloud/systemd"
)

var Domain = relations.Query{
	Label:      "Domain",
	Collection: "domains",
	Project: []relations.Project{{
		Key:   "name",
		Label: "Name",
	}, {
		Key:   "type",
		Label: "Type",
	}, {
		Key:   "root_domain",
		Label: "Root Domain",
	}},
	Relations: []relations.Relation{{
		Key:          "records",
		Label:        "Record",
		From:         "domains_records",
		LocalField:   "_id",
		ForeignField: "domain",
		BlockDelete:  true,
		Sort: map[string]int{
			"sub_domain": 1,
		},
		Project: []relations.Project{{
			Key: "deployment",
		}, {
			Key:   "sub_domain",
			Label: "Sub Domain",
		}, {
			Key:   "type",
			Label: "Type",
		}, {
			Key:   "value",
			Label: "Value",
		}, {
			Key:   "select",
			Label: "Select",
			Format: func(vals ...any) any {
				val := vals[0].(string)

				if val != "" {
					return val
				}

				return "-"
			},
		}, {
			Key:   "timestamp",
			Label: "Age",
			Format: func(vals ...any) any {
				val := vals[0]

				if mongoTime, ok := val.(bson.DateTime); ok {
					valTime := mongoTime.Time()
					return systemd.FormatUptimeShort(valTime)
				}

				if goTime, ok := val.(time.Time); ok {
					return systemd.FormatUptimeShort(goTime)
				}

				return "-"
			},
		}},
		Relations: []relations.Relation{{
			Key:          "deployments",
			Label:        "Deployment",
			From:         "deployments",
			LocalField:   "deployment",
			ForeignField: "_id",
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
				Label: "Age",
				Format: func(vals ...any) any {
					val := vals[0]

					if mongoTime, ok := val.(bson.DateTime); ok {
						valTime := mongoTime.Time()
						return systemd.FormatUptimeShort(valTime)
					}

					if goTime, ok := val.(time.Time); ok {
						return systemd.FormatUptimeShort(goTime)
					}

					return "-"
				},
			}},
		}},
	}},
}

func init() {
	relations.Register("domain", Domain)
}

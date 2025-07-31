package definitions

import (
	"time"

	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/relations"
	"github.com/pritunl/pritunl-cloud/systemd"
	"github.com/pritunl/pritunl-cloud/vm"
)

var Vpc = relations.Query{
	Label:      "VPC",
	Collection: "vpcs",
	Project: []relations.Project{{
		Key:   "name",
		Label: "Name",
	}, {
		Key:   "roles",
		Label: "Roles",
	}},
	Relations: []relations.Relation{{
		Key:          "instances",
		Label:        "Instance",
		From:         "instances",
		LocalField:   "_id",
		ForeignField: "vpc",
		BlockDelete:  true,
		Sort: map[string]int{
			"name": 1,
		},
		Project: []relations.Project{{
			Key:   "name",
			Label: "Name",
		}, {
			Keys: []string{
				"action",
				"state",
			},
			Label: "Status",
			Format: func(vals ...any) any {
				action, _ := vals[0].(string)
				state, _ := vals[1].(string)

				switch action {
				case instance.Start:
					switch state {
					case vm.Starting:
						return "Starting"
					case vm.Running:
						return "Running"
					case vm.Stopped:
						return "Starting"
					case vm.Failed:
						return "Starting"
					case vm.Updating:
						return "Updating"
					case vm.Provisioning:
						return "Provisioning"
					case "":
						return "Provisioning"
					}
				case instance.Cleanup:
					switch state {
					case vm.Starting:
						return "Stopping"
					case vm.Running:
						return "Stopping"
					case vm.Stopped:
						return "Stopping"
					case vm.Failed:
						return "Stopping"
					case vm.Updating:
						return "Updating"
					case vm.Provisioning:
						return "Stopping"
					case "":
						return "Stopping"
					}
				case instance.Stop:
					switch state {
					case vm.Starting:
						return "Stopping"
					case vm.Running:
						return "Stopping"
					case vm.Stopped:
						return "Stopped"
					case vm.Failed:
						return "Failed"
					case vm.Updating:
						return "Updating"
					case vm.Provisioning:
						return "Stopped"
					case "":
						return "Stopped"
					}
				case instance.Restart:
					return "Restarting"
				case instance.Destroy:
					return "Destroying"
				}

				return state
			},
		}, {
			Keys: []string{
				"timestamp",
				"action",
				"state",
			},
			Label: "Uptime",
			Format: func(vals ...any) any {
				val := vals[0]
				action, _ := vals[1].(string)
				state, _ := vals[2].(string)
				isActive := action == instance.Start ||
					state == vm.Running || state == vm.Starting ||
					state == vm.Provisioning

				if !isActive {
					return "-"
				}

				if mongoTime, ok := val.(primitive.DateTime); ok {
					valTime := mongoTime.Time()
					return systemd.FormatUptimeShort(valTime)
				}

				if goTime, ok := val.(time.Time); ok {
					return systemd.FormatUptimeShort(goTime)
				}

				return "-"
			},
		}, {
			Key:   "processors",
			Label: "Processors",
		}, {
			Key:   "memory",
			Label: "Memory",
		}, {
			Key:   "private_ips",
			Label: "Private IPv4",
		}, {
			Key:   "public_ips",
			Label: "Public IPv4",
		}},
	}},
}

func init() {
	relations.Register("vpc", Vpc)
}

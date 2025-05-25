package definitions

import (
	"time"

	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/relations"
	"github.com/pritunl/pritunl-cloud/systemd"
	"github.com/pritunl/pritunl-cloud/vm"
)

var Firewall = relations.Query{
	Label:      "Firewall",
	Collection: "firewalls",
	Project: []relations.Project{{
		Key:   "name",
		Label: "Name",
	}, {
		Key:   "network_roles",
		Label: "Roles",
	}},
	Relations: []relations.Relation{{
		Key:          "instances",
		Label:        "Instance",
		From:         "instances",
		LocalField:   "roles",
		ForeignField: "roles",
		Sort: map[string]int{
			"name": 1,
		},
		Project: []relations.Project{{
			Key:   "name",
			Label: "Name",
		}, {
			Keys: []string{
				"action",
				"virt_state",
			},
			Label: "Status",
			Format: func(vals ...any) any {
				action, _ := vals[0].(string)
				virtState, _ := vals[1].(string)

				switch action {
				case instance.Start:
					switch virtState {
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
					switch virtState {
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
					switch virtState {
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

				return virtState
			},
		}, {
			Keys: []string{
				"virt_timestamp",
				"action",
				"virt_state",
			},
			Label: "Uptime",
			Format: func(vals ...any) any {
				val := vals[0]
				action, _ := vals[1].(string)
				virtState, _ := vals[2].(string)
				isActive := action == instance.Start ||
					virtState == vm.Running || virtState == vm.Starting ||
					virtState == vm.Provisioning

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
		}},
	}},
}

func init() {
	relations.Register("firewall", Firewall)
}

package definitions

import (
	"time"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/relations"
	"github.com/pritunl/pritunl-cloud/systemd"
	"github.com/pritunl/pritunl-cloud/vm"
)

var Disk = relations.Query{
	Label:      "Disk",
	Collection: "disks",
	Project: []relations.Project{{
		Key:   "name",
		Label: "Name",
	}, {
		Key: "instance",
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
		Key:   "created",
		Label: "Created",
	}, {
		Key:   "type",
		Label: "Type",
	}, {
		Key:   "system_type",
		Label: "System Type",
	}, {
		Key:   "system_kind",
		Label: "System Kind",
	}, {
		Key:   "file_system",
		Label: "File System",
	}, {
		Key:   "uuid",
		Label: "UUID",
	}, {
		Key:   "delete_protection",
		Label: "Delete Protection",
	}, {
		Key:   "index",
		Label: "Index",
	}, {
		Key:   "size",
		Label: "Size",
	}, {
		Key:   "lv_size",
		Label: "LV Size",
	}},
	Relations: []relations.Relation{{
		Key:          "instance",
		Label:        "Instance",
		From:         "instances",
		LocalField:   "instance",
		ForeignField: "_id",
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

				if mongoTime, ok := val.(bson.DateTime); ok {
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
	relations.Register("disk", Disk)
}

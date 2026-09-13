package instances

import (
	"fmt"
	"image/color"
	"strings"
	"time"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/cli/widget"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/vpc"
	"github.com/sirupsen/logrus"
)

const (
	timeFormat = "2006-01-02 15:04:05"
)

// Item is an instance card with the referenced names resolved.
type Item struct {
	inst *instance.Instance
	vpcs []*vpc.Vpc
	node string
	zone string
	org  string
}

func (i *Item) Instance() *instance.Instance {
	return i.inst
}

func (i *Item) Id() string {
	return i.inst.Id.Hex()
}

func (i *Item) Name() string {
	return i.inst.Name
}

func (i *Item) Tag() string {
	return i.inst.Id.Hex()
}

// statusColor matches the status colors of the web instance rows.
func statusColor(status string) color.Color {
	if strings.Contains(status, "Restart Required") {
		return widget.ColorYellow
	}

	switch status {
	case "Running":
		return widget.ColorGreen
	case "Stopped", "Failed", "Destroying":
		return widget.ColorRed
	}

	return nil
}

func (i *Item) publicIp() string {
	if len(i.inst.PublicIps) > 0 {
		return i.inst.PublicIps[0]
	}
	if len(i.inst.HostIps) > 0 {
		return i.inst.HostIps[0]
	}
	return ""
}

func (i *Item) privateIp() string {
	if len(i.inst.PrivateIps) > 0 {
		return i.inst.PrivateIps[0]
	}
	return ""
}

func (i *Item) guestActive() bool {
	return i.inst.Guest != nil && i.inst.IsActive()
}

func (i *Item) load() string {
	if !i.guestActive() {
		return ""
	}
	return fmt.Sprintf("%.0f%% %.0f%% %.0f%%",
		i.inst.Guest.Load1, i.inst.Guest.Load5, i.inst.Guest.Load15)
}

func (i *Item) memory() string {
	if !i.guestActive() {
		return ""
	}
	return fmt.Sprintf("%.0f%%", i.inst.Guest.Memory)
}

func (i *Item) Fields() []resource.Field {
	return []resource.Field{
		{
			Label: "Status",
			Value: i.inst.Status,
			Color: statusColor(i.inst.Status),
		},
		{
			Label: "Uptime",
			Value: i.inst.Uptime,
		},
		{
			Label: "Node",
			Value: widget.Default(i.node, idHex(i.inst.Node)),
		},
		{
			Label: "Zone",
			Value: widget.Default(i.zone, idHex(i.inst.Zone)),
		},
		{
			Label: "Public IPv4",
			Value: i.publicIp(),
		},
		{
			Label: "Private IPv4",
			Value: i.privateIp(),
		},
		{
			Label: "Load",
			Value: i.load(),
		},
		{
			Label: "Memory",
			Value: i.memory(),
		},
	}
}

func idHex(id bson.ObjectID) string {
	if id.IsZero() {
		return ""
	}
	return id.Hex()
}

func formatTime(timestamp time.Time) string {
	if timestamp.IsZero() {
		return ""
	}
	return timestamp.Local().Format(timeFormat)
}

func (i *Item) nodePorts() string {
	ports := []string{}
	for _, mapping := range i.inst.NodePorts {
		ports = append(ports, fmt.Sprintf("%s %d -> %d",
			mapping.Protocol, mapping.ExternalPort, mapping.InternalPort))
	}
	return strings.Join(ports, ", ")
}

// Info mirrors the fields of the detailed instance view in the web
// interface, shown when the card is expanded.
func (i *Item) Info() []widget.InfoField {
	inst := i.inst

	info := inst.Info
	if info == nil {
		info = &instance.Info{}
	}

	nodeName := info.Node
	if nodeName == "" {
		nodeName = widget.Default(i.node, idHex(inst.Node))
	}

	mtu := ""
	if info.Mtu != 0 {
		mtu = fmt.Sprintf("%d", info.Mtu)
	}

	heartbeat := ""
	if inst.Guest != nil {
		heartbeat = formatTime(inst.Guest.Heartbeat)
	}

	fields := []widget.InfoField{
		{"ID", inst.Id.Hex()},
		{"Organization", widget.Default(i.org, idHex(inst.Organization))},
		{"Zone", widget.Default(i.zone, idHex(inst.Zone))},
		{"Node", nodeName},
		{"State", fmt.Sprintf("%s:%s",
			widget.Default(inst.Action, "None"),
			widget.Default(inst.State, "None"))},
		{"Status", inst.Status},
		{"Uptime", inst.Uptime},
		{"Comment", inst.Comment},
		{"Network Roles", strings.Join(inst.Roles, ", ")},
		{"Processors", fmt.Sprintf("%d", inst.Processors)},
		{"Memory", fmt.Sprintf("%d MB", inst.Memory)},
		{"Private IPv4", strings.Join(inst.PrivateIps, ", ")},
		{"Private IPv6", strings.Join(inst.PrivateIps6, ", ")},
		{"Public IPv4", strings.Join(inst.PublicIps, ", ")},
		{"Public IPv6", strings.Join(inst.PublicIps6, ", ")},
	}

	if len(inst.CloudPrivateIps) > 0 || len(inst.CloudPublicIps) > 0 ||
		len(inst.CloudPublicIps6) > 0 {

		fields = append(fields,
			widget.InfoField{"Cloud Private IPv4",
				strings.Join(inst.CloudPrivateIps, ", ")},
			widget.InfoField{"Cloud Public IPv4",
				strings.Join(inst.CloudPublicIps, ", ")},
			widget.InfoField{"Cloud Public IPv6",
				strings.Join(inst.CloudPublicIps6, ", ")},
		)
	}

	fields = append(fields,
		widget.InfoField{"Host IPv4", strings.Join(inst.HostIps, ", ")},
		widget.InfoField{"Gateway IPv4", strings.Join(inst.GatewayIps, ", ")},
		widget.InfoField{"Gateway IPv6", strings.Join(inst.GatewayIps6, ", ")},
		widget.InfoField{"Public MAC Address", inst.PublicMac},
		widget.InfoField{"Network MTU", mtu},
		widget.InfoField{"Network Namespace", inst.NetworkNamespace},
		widget.InfoField{"Node Ports", i.nodePorts()},
		widget.InfoField{"QEMU Version",
			widget.Default(inst.QemuVersion, "Unknown")},
		widget.InfoField{"Platform", widget.Bool(inst.Uefi, "UEFI", "BIOS")},
		widget.InfoField{"SecureBoot",
			widget.Bool(inst.SecureBoot, "Enabled", "Disabled")},
		widget.InfoField{"TPM", widget.Bool(inst.Tpm, "Enabled", "Disabled")},
		widget.InfoField{"Delete Protection",
			widget.Bool(inst.DeleteProtection, "Enabled", "Disabled")},
		widget.InfoField{"Disks", strings.Join(info.Disks, ", ")},
		widget.InfoField{"Authorities", strings.Join(info.Authorities, ", ")},
		widget.InfoField{"Agent Heartbeat", heartbeat},
		widget.InfoField{"Load", i.load()},
		widget.InfoField{"Memory Usage", i.memory()},
		widget.InfoField{"Deployment", idHex(inst.Deployment)},
		widget.InfoField{"Created", formatTime(inst.Created)},
	)

	if inst.Vnc {
		vncPort := ""
		if inst.VncDisplay != 0 {
			vncPort = fmt.Sprintf("%d", 5900+inst.VncDisplay)
		}
		fields = append(fields,
			widget.InfoField{"VNC IP", info.NodePublicIp},
			widget.InfoField{"VNC Port", vncPort},
		)
	}

	if inst.Spice {
		fields = append(fields,
			widget.InfoField{"Spice IP", info.NodePublicIp},
			widget.InfoField{"Spice Port", fmt.Sprintf("%d", inst.SpicePort)},
		)
	}

	return fields
}

// setAction applies the instance action the same way as the admin
// instances handler.
func setAction(instId bson.ObjectID, action string) func(
	db *database.Database) error {

	return func(db *database.Database) (err error) {
		doc := bson.M{
			"action": action,
		}
		if action != instance.Start {
			doc["restart"] = false
			doc["restart_block_ip"] = false
		}

		err = instance.UpdateMulti(db, []bson.ObjectID{instId}, &doc)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"instance_id": instId.Hex(),
			"action":      action,
		}).Info("tui: Instance action")

		err = event.PublishDispatch(db, "instance.change")
		if err != nil {
			return
		}

		return
	}
}

// deleteInstance queues the instance destroy the same way as the admin
// instance handler, instances with delete protection are refused.
func deleteInstance(instId bson.ObjectID) func(db *database.Database) error {
	return func(db *database.Database) (err error) {
		inst, err := instance.Get(db, instId)
		if err != nil {
			return
		}

		if inst.DeleteProtection {
			err = resource.DeleteError("instance",
				"Cannot delete instance with delete protection")
			return
		}

		return resource.Delete("instance", "instance.change", instId,
			instance.Delete)(db)
	}
}

// Actions mirrors the power and delete buttons of the detailed instance
// view, start is shown for stopped instances, stop and restart for
// running.
func (i *Item) Actions() []resource.Action {
	inst := i.inst
	actions := []resource.Action{}

	if inst.Action == instance.Stop {
		actions = append(actions, resource.Action{
			Key:     "s",
			Label:   "Start",
			Status:  "Starting",
			Confirm: fmt.Sprintf("Start the instance %s?", inst.Name),
			Run:     setAction(inst.Id, instance.Start),
		})
	}

	if inst.Action == instance.Start || inst.Action == instance.Restart {
		actions = append(actions, resource.Action{
			Key:     "x",
			Label:   "Stop",
			Status:  "Stopping",
			Danger:  true,
			Confirm: fmt.Sprintf("Stop the instance %s?", inst.Name),
			Run:     setAction(inst.Id, instance.Stop),
		})
	}

	if inst.Action == instance.Start {
		actions = append(actions, resource.Action{
			Key:     "r",
			Label:   "Restart",
			Status:  "Restarting",
			Danger:  true,
			Confirm: fmt.Sprintf("Restart the instance %s?", inst.Name),
			Run:     setAction(inst.Id, instance.Restart),
		})
	}

	actions = append(actions, resource.DeleteAction("instance", inst.Name,
		deleteInstance(inst.Id)))

	return actions
}

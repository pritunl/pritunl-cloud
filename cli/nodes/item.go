package nodes

import (
	"fmt"
	"image/color"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/cli/widget"
	"github.com/pritunl/pritunl-cloud/node"
)

const (
	timeFormat = "2006-01-02 15:04:05"
)

var typeLabels = map[string]string{
	node.Admin:      "Admin",
	node.User:       "User",
	node.Balancer:   "Load Balancer",
	node.Hypervisor: "Hypervisor",
}

// Item is a node card with the zone name resolved.
type Item struct {
	nde   *node.Node
	names *names
	zone  string
}

func (i *Item) Node() *node.Node {
	return i.nde
}

func (i *Item) Id() string {
	return i.nde.Id.Hex()
}

func (i *Item) Name() string {
	return i.nde.Name
}

func (i *Item) Tag() string {
	return i.nde.Id.Hex()
}

// active mirrors the web node row, a node with no requests, memory or
// load is shown as inactive.
func (i *Item) active() bool {
	nde := i.nde
	if nde.RequestsMin != 0 || nde.Memory != 0 {
		return true
	}
	if nde.Metric != nil && (nde.Metric.Load1 != 0 ||
		nde.Metric.Load5 != 0 || nde.Metric.Load15 != 0) {

		return true
	}
	return false
}

func (i *Item) status() (string, color.Color) {
	if !i.active() {
		return "Inactive", widget.ColorRed
	}
	return "Active " + formatTime(i.nde.Timestamp), widget.ColorGreen
}

func formatTime(timestamp time.Time) string {
	if timestamp.IsZero() {
		return ""
	}
	return timestamp.Local().Format(timeFormat)
}

func idHex(id bson.ObjectID) string {
	if id.IsZero() {
		return ""
	}
	return id.Hex()
}

func (i *Item) types() string {
	labels := []string{}
	for _, typ := range i.nde.Types {
		label := typeLabels[typ]
		if label == "" {
			label = typ
		}
		labels = append(labels, label)
	}
	return strings.Join(labels, ", ")
}

func (i *Item) load() string {
	if i.nde.Metric == nil || !i.active() {
		return ""
	}
	return fmt.Sprintf("%.0f%% %.0f%% %.0f%%",
		i.nde.Metric.Load1, i.nde.Metric.Load5, i.nde.Metric.Load15)
}

func (i *Item) memory() string {
	if i.nde.Metric == nil || !i.active() {
		return ""
	}
	if i.nde.Hugepages {
		return fmt.Sprintf("%.0f%% (hugepages %.0f%%)",
			i.nde.Metric.Memory, i.nde.Metric.HugePages)
	}
	return fmt.Sprintf("%.0f%%", i.nde.Metric.Memory)
}

func (i *Item) privateIps() string {
	ips := []string{}
	for _, ip := range i.nde.PrivateIps {
		ips = append(ips, ip)
	}
	sort.Strings(ips)
	return strings.Join(ips, ", ")
}

func (i *Item) Fields() []resource.Field {
	status, statusColor := i.status()

	return []resource.Field{
		{
			Label: "Status",
			Value: status,
			Color: statusColor,
		},
		{
			Label: "Types",
			Value: i.types(),
		},
		{
			Label: "Zone",
			Value: widget.Default(i.zone, idHex(i.nde.Zone)),
		},
		{
			Label: "Requests",
			Value: fmt.Sprintf("%d/min", i.nde.RequestsMin),
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

// Info mirrors the fields of the detailed node view in the web interface.
func (i *Item) Info() []widget.InfoField {
	nde := i.nde

	timestamp := formatTime(nde.Timestamp)
	if !i.active() {
		timestamp = "Inactive"
	}

	cpuUnits := "Unknown"
	if nde.CpuUnits != 0 {
		cpuUnits = fmt.Sprintf("%d", nde.CpuUnits)
	}
	memoryUnits := "Unknown"
	if nde.MemoryUnits != 0 {
		memoryUnits = fmt.Sprintf("%d", int(math.Floor(nde.MemoryUnits)))
	}

	fields := []widget.InfoField{
		{"ID", nde.Id.Hex()},
		{"Version", widget.Default(nde.SoftwareVersion, "Unknown")},
		{"Timestamp", timestamp},
		{"Types", i.types()},
		{"Zone", widget.Default(i.zone, idHex(nde.Zone))},
		{"CPU Units Reserved", fmt.Sprintf("%d", nde.CpuUnitsRes)},
		{"CPU Units", cpuUnits},
		{"Memory Units Reserved", fmt.Sprintf("%.0f", nde.MemoryUnitsRes)},
		{"Memory Units", memoryUnits},
		{"Default Interface", widget.Default(nde.DefaultInterface, "Unknown")},
		{"Hostname", widget.Default(nde.Hostname, "Unknown")},
		{"Private IPv4", i.privateIps()},
		{"Public IPv4", strings.Join(nde.PublicIps, ", ")},
		{"Public IPv6", strings.Join(nde.PublicIps6, ", ")},
		{"Requests", fmt.Sprintf("%d/min", nde.RequestsMin)},
		{"Load", i.load()},
		{"Memory Usage", i.memory()},
	}

	if nde.OraclePublicKey != "" {
		fields = append(fields, widget.InfoField{
			Label: "Oracle Cloud Public Key",
			Value: nde.OraclePublicKey,
		})
	}

	return fields
}

// Actions mirrors the delete button of the detailed node view, the web
// interface disables delete while the node is active. Node restarts are
// not exposed.
func (i *Item) Actions() []resource.Action {
	if i.active() {
		return nil
	}

	return []resource.Action{
		resource.DeleteAction("node", i.nde.Name,
			resource.DeleteRelated("node", "node", "node.change", i.nde.Id,
				node.Remove)),
	}
}

func (i *Item) Editor() resource.Editor {
	return newEditor(i.nde, i.names)
}

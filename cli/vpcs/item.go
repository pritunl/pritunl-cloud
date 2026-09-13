package vpcs

import (
	"fmt"

	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/cli/widget"
	"github.com/pritunl/pritunl-cloud/vpc"
)

// Item is a VPC card with the organization and datacenter names
// resolved.
type Item struct {
	vc         *vpc.Vpc
	names      *names
	org        string
	datacenter string
}

func (i *Item) Vpc() *vpc.Vpc {
	return i.vc
}

func (i *Item) Id() string {
	return i.vc.Id.Hex()
}

func (i *Item) Name() string {
	return i.vc.Name
}

func (i *Item) Tag() string {
	return i.vc.Id.Hex()
}

func (i *Item) organization() string {
	return widget.Default(i.org, resource.IdHex(i.vc.Organization))
}

func (i *Item) datacenterName() string {
	return widget.Default(i.datacenter, resource.IdHex(i.vc.Datacenter))
}

func (i *Item) vlan() string {
	if i.vc.VpcId == 0 {
		return ""
	}
	return fmt.Sprintf("%d", i.vc.VpcId)
}

func (i *Item) Fields() []resource.Field {
	return []resource.Field{
		{
			Label: "Organization",
			Value: i.organization(),
		},
		{
			Label: "Datacenter",
			Value: i.datacenterName(),
		},
		{
			Label: "Network",
			Value: i.vc.Network,
		},
		{
			Label: "VLAN",
			Value: i.vlan(),
		},
	}
}

// Info mirrors the fields of the detailed VPC view.
func (i *Item) Info() []widget.InfoField {
	return []widget.InfoField{
		{"ID", i.vc.Id.Hex()},
		{"Datacenter", i.datacenterName()},
		{"Organization", i.organization()},
		{"VLAN Number", widget.Default(i.vlan(), "Unknown")},
		{"Private IPv6 Network", widget.Default(i.vc.Network6, "Unknown")},
	}
}

// Actions mirrors the delete button of the detailed view, the admin
// handler checks the relations before removing the VPC.
func (i *Item) Actions() []resource.Action {
	return []resource.Action{
		resource.DeleteAction("VPC", i.vc.Name,
			resource.DeleteRelated("vpc", "VPC", "vpc.change", i.vc.Id,
				vpc.Remove)),
	}
}

func (i *Item) Editor() resource.Editor {
	return newEditor(i.vc, i.names, false)
}

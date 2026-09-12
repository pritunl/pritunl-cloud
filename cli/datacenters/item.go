package datacenters

import (
	"fmt"

	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/cli/widget"
	"github.com/pritunl/pritunl-cloud/datacenter"
)

// Item is a datacenter card.
type Item struct {
	dc    *datacenter.Datacenter
	names *names
}

func (i *Item) Datacenter() *datacenter.Datacenter {
	return i.dc
}

func (i *Item) Id() string {
	return i.dc.Id.Hex()
}

func (i *Item) Name() string {
	return i.dc.Name
}

func (i *Item) Tag() string {
	return i.dc.Id.Hex()
}

// networkModeLabel returns the display name of the network mode.
func networkModeLabel(mode string) string {
	for _, opt := range networkModeOptions {
		if opt.Value == mode {
			return opt.Label
		}
	}
	return mode
}

func (i *Item) Fields() []resource.Field {
	return []resource.Field{
		{
			Label: "Network Mode",
			Value: networkModeLabel(i.dc.NetworkMode),
		},
		{
			Label: "Public Storages",
			Value: fmt.Sprintf("%d", len(i.dc.PublicStorages)),
		},
		{
			Label: "Comment",
			Value: i.dc.Comment,
		},
	}
}

// Info mirrors the fields of the detailed datacenter view.
func (i *Item) Info() []widget.InfoField {
	return []widget.InfoField{
		{"ID", i.dc.Id.Hex()},
	}
}

// Actions mirrors the delete button of the detailed view, the admin
// handler checks the relations before removing the datacenter.
func (i *Item) Actions() []resource.Action {
	return []resource.Action{
		resource.DeleteAction("datacenter", i.dc.Name,
			resource.DeleteRelated("datacenter", "datacenter", "datacenter.change", i.dc.Id,
				datacenter.Remove)),
	}
}

func (i *Item) Editor() resource.Editor {
	return newEditor(i.dc, i.names)
}

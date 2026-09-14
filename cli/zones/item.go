package zones

import (
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/cli/widget"
	"github.com/pritunl/pritunl-cloud/zone"
)

// Item is a zone card with the datacenter name resolved.
type Item struct {
	zne        *zone.Zone
	datacenter string
}

func (i *Item) Zone() *zone.Zone {
	return i.zne
}

func (i *Item) Id() string {
	return i.zne.Id.Hex()
}

func (i *Item) Name() string {
	return i.zne.Name
}

func (i *Item) Tag() string {
	return i.zne.Id.Hex()
}

func (i *Item) datacenterName() string {
	return widget.Default(i.datacenter, resource.IdHex(i.zne.Datacenter))
}

func (i *Item) Fields() []resource.Field {
	return []resource.Field{
		{
			Label: "Datacenter",
			Value: i.datacenterName(),
		},
		{
			Label: "Comment",
			Value: i.zne.Comment,
		},
	}
}

// Info mirrors the fields of the detailed zone view.
func (i *Item) Info() []widget.InfoField {
	return []widget.InfoField{
		{"ID", i.zne.Id.Hex()},
		{"Datacenter", i.datacenterName()},
	}
}

// Actions mirrors the delete button of the detailed view, the admin
// handler checks the relations before removing the zone.
func (i *Item) Actions() []resource.Action {
	return []resource.Action{
		resource.DeleteAction("zone", i.zne.Name,
			resource.DeleteRelated("zone", "zone", "zone.change", i.zne.Id,
				zone.Remove)),
	}
}

func (i *Item) Editor() resource.Editor {
	return newEditor(i.zne, nil)
}

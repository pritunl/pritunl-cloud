package blocks

import (
	"fmt"

	"github.com/pritunl/pritunl-cloud/aggregate"
	"github.com/pritunl/pritunl-cloud/block"
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/cli/widget"
)

// Item is a block card with the address counts of the block.
type Item struct {
	blck *aggregate.BlockAggregate
}

func (i *Item) Block() *block.Block {
	return &i.blck.Block
}

func (i *Item) Id() string {
	return i.blck.Id.Hex()
}

func (i *Item) Name() string {
	return i.blck.Name
}

func (i *Item) Tag() string {
	return i.blck.Id.Hex()
}

// gateway returns the IPv4 or IPv6 gateway like the web block rows.
func (i *Item) gateway() string {
	return widget.Default(i.blck.Gateway, i.blck.Gateway6)
}

func typeLabel(typ string) string {
	for _, opt := range typeOptions {
		if opt.Value == typ {
			return opt.Label
		}
	}
	return typ
}

func (i *Item) Fields() []resource.Field {
	return []resource.Field{
		{
			Label: "Network Mode",
			Value: typeLabel(i.blck.Type),
		},
		{
			Label: "Gateway",
			Value: i.gateway(),
		},
		{
			Label: "Available",
			Value: fmt.Sprintf("%d", i.blck.Available),
		},
		{
			Label: "Capacity",
			Value: fmt.Sprintf("%d", i.blck.Capacity),
		},
	}
}

// Info mirrors the fields of the detailed block view.
func (i *Item) Info() []widget.InfoField {
	return []widget.InfoField{
		{"ID", i.blck.Id.Hex()},
		{"Available", fmt.Sprintf("%d", i.blck.Available)},
		{"Capacity", fmt.Sprintf("%d", i.blck.Capacity)},
	}
}

// Actions mirrors the delete button of the detailed view, the admin
// handler checks the relations before removing the block.
func (i *Item) Actions() []resource.Action {
	return []resource.Action{
		resource.DeleteAction("block", i.blck.Name,
			resource.DeleteRelated("block", "block", "block.change", i.blck.Id,
				block.Remove)),
	}
}

func (i *Item) Editor() resource.Editor {
	return newEditor(i.blck)
}

package pools

import (
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/cli/widget"
	"github.com/pritunl/pritunl-cloud/pool"
)

// Item is a pool card with the zone and datacenter names resolved.
type Item struct {
	pl         *pool.Pool
	names      *names
	zone       string
	datacenter string
}

func (i *Item) Pool() *pool.Pool {
	return i.pl
}

func (i *Item) Id() string {
	return i.pl.Id.Hex()
}

func (i *Item) Name() string {
	return i.pl.Name
}

func (i *Item) Tag() string {
	return i.pl.Id.Hex()
}

func (i *Item) Fields() []resource.Field {
	return []resource.Field{
		{
			Label: "Zone",
			Value: widget.Default(i.zone, resource.IdHex(i.pl.Zone)),
		},
		{
			Label: "Volume Group",
			Value: i.pl.VgName,
		},
		{
			Label: "Comment",
			Value: i.pl.Comment,
		},
	}
}

// Info mirrors the fields of the detailed pool view.
func (i *Item) Info() []widget.InfoField {
	return []widget.InfoField{
		{"ID", i.pl.Id.Hex()},
	}
}

// Actions mirrors the delete button of the detailed pool view.
func (i *Item) Actions() []resource.Action {
	return []resource.Action{
		resource.DeleteAction("pool", i.pl.Name,
			resource.Delete("pool", "pool.change", i.pl.Id, pool.Remove)),
	}
}

func (i *Item) Editor() resource.Editor {
	return newEditor(i.pl, i.names)
}

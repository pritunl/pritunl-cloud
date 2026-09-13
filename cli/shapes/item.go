package shapes

import (
	"fmt"
	"strings"

	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/cli/widget"
	"github.com/pritunl/pritunl-cloud/shape"
)

// Item is a shape card.
type Item struct {
	shpe *shape.Shape
}

func (i *Item) Shape() *shape.Shape {
	return i.shpe
}

func (i *Item) Id() string {
	return i.shpe.Id.Hex()
}

func (i *Item) Name() string {
	return i.shpe.Name
}

func (i *Item) Tag() string {
	return i.shpe.Id.Hex()
}

func (i *Item) Fields() []resource.Field {
	return []resource.Field{
		{
			Label: "Memory",
			Value: fmt.Sprintf("%d MB", i.shpe.Memory),
		},
		{
			Label: "Processors",
			Value: fmt.Sprintf("%d", i.shpe.Processors),
		},
		{
			Label: "Flexible",
			Value: widget.Bool(i.shpe.Flexible, "Yes", "No"),
		},
		{
			Label: "Roles",
			Value: strings.Join(i.shpe.Roles, ", "),
		},
	}
}

// Info mirrors the fields of the detailed shape view.
func (i *Item) Info() []widget.InfoField {
	return []widget.InfoField{
		{"ID", i.shpe.Id.Hex()},
		{"Node Count", fmt.Sprintf("%d", i.shpe.NodeCount)},
	}
}

// Actions mirrors the delete button of the detailed view, the admin
// handler checks the relations before removing the shape.
func (i *Item) Actions() []resource.Action {
	return []resource.Action{
		resource.DeleteAction("shape", i.shpe.Name,
			resource.DeleteRelated("shape", "shape", "shape.change", i.shpe.Id,
				shape.Remove)),
	}
}

func (i *Item) Editor() resource.Editor {
	return newEditor(i.shpe)
}

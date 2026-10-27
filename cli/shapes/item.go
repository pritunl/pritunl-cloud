package shapes

import (
	"fmt"
	"strings"

	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/cli/widget"
	"github.com/pritunl/pritunl-cloud/shape"
)

// Item is a shape card with the pool choices of the settings form.
type Item struct {
	shpe  *shape.Shape
	pools []widget.SelectOption
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

// poolNames returns the labels of the disk pools of the shape.
func (i *Item) poolNames() string {
	names := []string{}
	for _, poolId := range i.shpe.DiskPools {
		name := poolId.Hex()
		for _, opt := range i.pools {
			if opt.Value == name {
				name = opt.Label
				break
			}
		}
		names = append(names, name)
	}
	return strings.Join(names, ", ")
}

// Fields shows the resources of an instance shape or the storage of a
// disk shape like the shape cards of the web interface.
func (i *Item) Fields() []resource.Field {
	if i.shpe.GetType() == shape.Disk {
		storageLabel := "Disks Path"
		storageValue := i.shpe.DisksPath
		diskType := "QCOW"
		if i.shpe.DiskType == shape.Lvm {
			storageLabel = "Disk Pools"
			storageValue = i.poolNames()
			diskType = "LVM"
		}

		return []resource.Field{
			{
				Label: "Type",
				Value: "Disk",
			},
			{
				Label: "Disk Type",
				Value: diskType,
			},
			{
				Label: storageLabel,
				Value: storageValue,
			},
			{
				Label: "Roles",
				Value: strings.Join(i.shpe.Roles, ", "),
			},
		}
	}

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
		{"Type", i.shpe.GetType()},
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
	return newEditor(i.shpe, i.pools)
}

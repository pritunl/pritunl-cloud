package organizations

import (
	"strings"

	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/cli/widget"
	"github.com/pritunl/pritunl-cloud/organization"
)

// Item is an organization card.
type Item struct {
	org *organization.Organization
}

func (i *Item) Organization() *organization.Organization {
	return i.org
}

func (i *Item) Id() string {
	return i.org.Id.Hex()
}

func (i *Item) Name() string {
	return i.org.Name
}

func (i *Item) Tag() string {
	return i.org.Id.Hex()
}

func (i *Item) Fields() []resource.Field {
	return []resource.Field{
		{
			Label: "Roles",
			Value: strings.Join(i.org.Roles, ", "),
		},
		{
			Label: "Comment",
			Value: i.org.Comment,
		},
	}
}

// Info mirrors the fields of the detailed organization view.
func (i *Item) Info() []widget.InfoField {
	return []widget.InfoField{
		{"ID", i.org.Id.Hex()},
	}
}

// Actions mirrors the delete button of the detailed organization view,
// the admin handler checks the relations before removing.
func (i *Item) Actions() []resource.Action {
	return []resource.Action{
		resource.DeleteAction("organization", i.org.Name,
			resource.DeleteRelated("organization", "organization",
				"organization.change", i.org.Id, organization.Remove)),
	}
}

func (i *Item) Editor() resource.Editor {
	return newEditor(i.org)
}

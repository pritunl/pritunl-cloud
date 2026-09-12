package authorities

import (
	"strings"

	"github.com/pritunl/pritunl-cloud/authority"
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/cli/widget"
	"github.com/pritunl/pritunl-cloud/database"
)

// Item is an authority card with the organization name resolved.
type Item struct {
	authr *authority.Authority
	orgs  []*database.Named
	org   string
}

func (i *Item) Authority() *authority.Authority {
	return i.authr
}

func (i *Item) Id() string {
	return i.authr.Id.Hex()
}

func (i *Item) Name() string {
	return i.authr.Name
}

func (i *Item) Tag() string {
	return i.authr.Id.Hex()
}

// organization returns the organization name or the label of the web
// authority rows for authorities without one.
func (i *Item) organization() string {
	if i.authr.Organization.IsZero() {
		return "No Organization"
	}
	return widget.Default(i.org, i.authr.Organization.Hex())
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
			Label: "Organization",
			Value: i.organization(),
		},
		{
			Label: "Type",
			Value: typeLabel(i.authr.Type),
		},
		{
			Label: "Roles",
			Value: strings.Join(i.authr.Roles, ", "),
		},
	}
}

// Info mirrors the fields of the detailed authority view.
func (i *Item) Info() []widget.InfoField {
	return []widget.InfoField{
		{"ID", i.authr.Id.Hex()},
	}
}

// Actions mirrors the delete button of the detailed view.
func (i *Item) Actions() []resource.Action {
	return []resource.Action{
		resource.DeleteAction("authority", i.authr.Name,
			resource.Delete("authority", "authority.change", i.authr.Id, authority.Remove)),
	}
}

func (i *Item) Editor() resource.Editor {
	return newEditor(i.authr, i.orgs)
}

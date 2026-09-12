package firewalls

import (
	"fmt"
	"strings"

	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/cli/widget"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/firewall"
)

// Item is a firewall card with the organization name resolved.
type Item struct {
	fire *firewall.Firewall
	orgs []*database.Named
	org  string
}

func (i *Item) Firewall() *firewall.Firewall {
	return i.fire
}

func (i *Item) Id() string {
	return i.fire.Id.Hex()
}

func (i *Item) Name() string {
	return i.fire.Name
}

func (i *Item) Tag() string {
	return i.fire.Id.Hex()
}

// organization returns the organization name or the node firewall label
// like the web firewall rows.
func (i *Item) organization() string {
	if i.fire.Organization.IsZero() {
		return nodeFirewall
	}
	return widget.Default(i.org, i.fire.Organization.Hex())
}

func (i *Item) Fields() []resource.Field {
	return []resource.Field{
		{
			Label: "Organization",
			Value: i.organization(),
		},
		{
			Label: "Roles",
			Value: strings.Join(i.fire.Roles, ", "),
		},
		{
			Label: "Ingress",
			Value: fmt.Sprintf("%d rules", len(i.fire.Ingress)),
		},
		{
			Label: "Comment",
			Value: i.fire.Comment,
		},
	}
}

// Info mirrors the fields of the detailed firewall view, the ingress
// rules are shown in the settings form.
func (i *Item) Info() []widget.InfoField {
	return []widget.InfoField{
		{"ID", i.fire.Id.Hex()},
		{"Organization", i.organization()},
		{"Roles", strings.Join(i.fire.Roles, ", ")},
		{"Comment", i.fire.Comment},
	}
}

// Actions mirrors the delete button of the detailed view, the admin
// handler checks the relations before removing the firewall.
func (i *Item) Actions() []resource.Action {
	return []resource.Action{
		resource.DeleteAction("firewall", i.fire.Name,
			resource.DeleteRelated("firewall", "firewall", "firewall.change", i.fire.Id,
				firewall.Remove)),
	}
}

func (i *Item) Editor() resource.Editor {
	return newEditor(i.fire, i.orgs)
}

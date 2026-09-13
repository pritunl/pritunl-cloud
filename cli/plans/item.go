package plans

import (
	"fmt"

	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/cli/widget"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/plan"
)

// Item is a plan card with the organization name resolved.
type Item struct {
	pln  *plan.Plan
	orgs []*database.Named
	org  string
}

func (i *Item) Plan() *plan.Plan {
	return i.pln
}

func (i *Item) Id() string {
	return i.pln.Id.Hex()
}

func (i *Item) Name() string {
	return i.pln.Name
}

func (i *Item) Tag() string {
	return i.pln.Id.Hex()
}

// organization returns the organization name like the web plan rows.
func (i *Item) organization() string {
	if i.pln.Organization.IsZero() {
		return "Unknown Organization"
	}
	return widget.Default(i.org, i.pln.Organization.Hex())
}

func (i *Item) Fields() []resource.Field {
	return []resource.Field{
		{
			Label: "Organization",
			Value: i.organization(),
		},
		{
			Label: "Statements",
			Value: fmt.Sprintf("%d", len(i.pln.Statements)),
		},
		{
			Label: "Comment",
			Value: i.pln.Comment,
		},
	}
}

// Info mirrors the fields of the detailed plan view.
func (i *Item) Info() []widget.InfoField {
	return []widget.InfoField{
		{"ID", i.pln.Id.Hex()},
	}
}

// Actions mirrors the delete button of the detailed plan view.
func (i *Item) Actions() []resource.Action {
	return []resource.Action{
		resource.DeleteAction("plan", i.pln.Name,
			resource.Delete("plan", "plan.change", i.pln.Id, plan.Remove)),
	}
}

func (i *Item) Editor() resource.Editor {
	return newEditor(i.pln, i.orgs)
}

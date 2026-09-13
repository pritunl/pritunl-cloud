package secrets

import (
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/cli/widget"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/secret"
)

// Item is a secret card with the organization name resolved.
type Item struct {
	secr *secret.Secret
	orgs []*database.Named
	org  string
}

func (i *Item) Secret() *secret.Secret {
	return i.secr
}

func (i *Item) Id() string {
	return i.secr.Id.Hex()
}

func (i *Item) Name() string {
	return i.secr.Name
}

func (i *Item) Tag() string {
	return i.secr.Id.Hex()
}

// organization returns the organization name or the node secret label
// like the web secret rows.
func (i *Item) organization() string {
	if i.secr.Organization.IsZero() {
		return nodeSecret
	}
	return widget.Default(i.org, i.secr.Organization.Hex())
}

func typeLabel(typ string) string {
	for _, opt := range typeOptions {
		if opt.Value == typ {
			return opt.Label
		}
	}
	return "Unknown"
}

func (i *Item) Fields() []resource.Field {
	return []resource.Field{
		{
			Label: "Organization",
			Value: i.organization(),
		},
		{
			Label: "Type",
			Value: typeLabel(i.secr.Type),
		},
		{
			Label: "Comment",
			Value: i.secr.Comment,
		},
	}
}

// Info mirrors the fields of the detailed secret view, the keys are
// only shown in the settings form.
func (i *Item) Info() []widget.InfoField {
	return []widget.InfoField{
		{"ID", i.secr.Id.Hex()},
	}
}

// Actions mirrors the delete button of the detailed view, the admin
// handler checks the relations before removing the secret.
func (i *Item) Actions() []resource.Action {
	return []resource.Action{
		resource.DeleteAction("secret", i.secr.Name,
			resource.DeleteRelated("secret", "secret", "secret.change", i.secr.Id,
				secret.Remove)),
	}
}

func (i *Item) Editor() resource.Editor {
	return newEditor(i.secr, i.orgs)
}

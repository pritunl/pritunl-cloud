package domains

import (
	"fmt"

	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/cli/widget"
	"github.com/pritunl/pritunl-cloud/domain"
)

// Item is a domain card with the organization name resolved.
type Item struct {
	domn  *domain.Domain
	names *names
	org   string
}

func (i *Item) Domain() *domain.Domain {
	return i.domn
}

func (i *Item) Id() string {
	return i.domn.Id.Hex()
}

func (i *Item) Name() string {
	return i.domn.Name
}

func (i *Item) Tag() string {
	return i.domn.Id.Hex()
}

// organization returns the organization name or the node domain label
// like the web domain rows.
func (i *Item) organization() string {
	if i.domn.Organization.IsZero() {
		return nodeDomain
	}
	return widget.Default(i.org, i.domn.Organization.Hex())
}

// providerLabel returns the provider name shown on the web domain rows,
// local domains have no provider.
func providerLabel(typ string) string {
	switch typ {
	case domain.AWS:
		return "AWS"
	case domain.Cloudflare:
		return "Cloudflare"
	case domain.OracleCloud:
		return "Oracle Cloud"
	}
	return ""
}

func (i *Item) Fields() []resource.Field {
	return []resource.Field{
		{
			Label: "Organization",
			Value: i.organization(),
		},
		{
			Label: "Provider",
			Value: providerLabel(i.domn.Type),
		},
		{
			Label: "Domain",
			Value: i.domn.RootDomain,
		},
		{
			Label: "Records",
			Value: fmt.Sprintf("%d", len(i.domn.Records)),
		},
	}
}

// Info mirrors the fields of the detailed domain view.
func (i *Item) Info() []widget.InfoField {
	return []widget.InfoField{
		{"ID", i.domn.Id.Hex()},
	}
}

// Actions mirrors the delete button of the detailed view, the admin
// handler checks the relations before removing the domain.
func (i *Item) Actions() []resource.Action {
	return []resource.Action{
		resource.DeleteAction("domain", i.domn.Name,
			resource.DeleteRelated("domain", "domain", "domain.change", i.domn.Id,
				domain.Remove)),
	}
}

func (i *Item) Editor() resource.Editor {
	return newEditor(i.domn, i.names)
}

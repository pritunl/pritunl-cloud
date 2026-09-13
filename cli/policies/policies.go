// Package policies provides the policies tab, a paged list of the user
// authentication policies in the cluster from the admin perspective.
package policies

import (
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/cli/widget"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/policy"
	"github.com/pritunl/pritunl-cloud/settings"
)

const (
	// cardFields is the number of fields on each policy card.
	cardFields = 3
)

type Provider struct{}

func New() *Provider {
	return &Provider{}
}

func (p *Provider) Title() string {
	return "Policies"
}

func (p *Provider) Empty() string {
	return "No policies"
}

func (p *Provider) Singular() string {
	return "policy"
}

func (p *Provider) FieldCount() int {
	return cardFields
}

// FilterFields mirrors the policies filter of the web interface.
func (p *Provider) FilterFields(db *database.Database) (
	fields []resource.FilterField, err error) {

	fields = []resource.FilterField{
		{
			Key:         "id",
			Label:       "Policy ID",
			Placeholder: "Policy ID",
		},
		{
			Key:         "name",
			Label:       "Name",
			Placeholder: "Name",
		},
		{
			Key:         "comment",
			Label:       "Comment",
			Placeholder: "Comment",
		},
	}

	return
}

// providerOptions lists the secondary authentication providers of the
// settings, the choices of the two-factor selects.
func providerOptions() []widget.SelectOption {
	opts := []widget.SelectOption{}
	for _, provider := range settings.Auth.SecondaryProviders {
		opts = append(opts, widget.SelectOption{
			Label: provider.Name,
			Value: provider.Id.Hex(),
		})
	}
	return opts
}

// NewEditor returns the settings form of a new policy.
func (p *Provider) NewEditor(db *database.Database) (
	editor resource.Editor, err error) {

	e := newEditor(&policy.Policy{
		Name:  "new-policy",
		Roles: []string{},
		Rules: map[string]*policy.Rule{},
	}, providerOptions())
	e.create = true
	editor = e

	return
}

// Load returns a page of policies with the two-factor providers of the
// settings as the select options.
func (p *Provider) Load(db *database.Database, filter resource.Filter,
	page, pageCount int64) (result *resource.Page, err error) {

	query := buildQuery(filter)

	policies, count, err := policy.GetAllPaged(db, &query, page, pageCount)
	if err != nil {
		return
	}

	providers := providerOptions()

	items := make([]resource.Item, 0, len(policies))
	for _, polcy := range policies {
		items = append(items, &Item{
			polcy:     polcy,
			providers: providers,
		})
	}

	result = &resource.Page{
		Items: items,
		Count: count,
	}

	return
}

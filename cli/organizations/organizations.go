// Package organizations provides the organizations tab, a paged list of
// the organizations in the cluster from the admin perspective.
package organizations

import (
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/organization"
)

const (
	// cardFields is the number of fields on each organization card.
	cardFields = 2
)

type Provider struct{}

func New() *Provider {
	return &Provider{}
}

func (p *Provider) Title() string {
	return "Organizations"
}

func (p *Provider) Empty() string {
	return "No organizations"
}

func (p *Provider) Singular() string {
	return "organization"
}

func (p *Provider) FieldCount() int {
	return cardFields
}

// FilterFields mirrors the organizations filter of the web interface.
func (p *Provider) FilterFields(db *database.Database) (
	fields []resource.FilterField, err error) {

	fields = []resource.FilterField{
		{
			Key:         "id",
			Label:       "Organization ID",
			Placeholder: "Organization ID",
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

// NewEditor returns the settings form of a new organization.
func (p *Provider) NewEditor(db *database.Database) (
	editor resource.Editor, err error) {

	e := newEditor(&organization.Organization{
		Name:  "new-organization",
		Roles: []string{},
	})
	e.create = true
	editor = e

	return
}

// Load returns a page of organizations.
func (p *Provider) Load(db *database.Database, filter resource.Filter,
	page, pageCount int64) (result *resource.Page, err error) {

	query := buildQuery(filter)

	orgs, count, err := organization.GetAllPaged(db, &query, page,
		pageCount)
	if err != nil {
		return
	}

	items := make([]resource.Item, 0, len(orgs))
	for _, org := range orgs {
		items = append(items, &Item{
			org: org,
		})
	}

	result = &resource.Page{
		Items: items,
		Count: count,
	}

	return
}

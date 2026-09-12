// Package authorities provides the authorities tab, a paged list of the
// SSH authorities in the cluster from the admin perspective.
package authorities

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/authority"
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/database"
)

const (
	// cardFields is the number of fields on each authority card.
	cardFields = 3
)

type Provider struct{}

func New() *Provider {
	return &Provider{}
}

func (p *Provider) Title() string {
	return "Authorities"
}

func (p *Provider) Empty() string {
	return "No authorities"
}

func (p *Provider) Singular() string {
	return "authority"
}

func (p *Provider) FieldCount() int {
	return cardFields
}

// FilterFields mirrors the authorities filter of the web interface.
func (p *Provider) FilterFields(db *database.Database) (
	fields []resource.FilterField, err error) {

	orgs, err := resource.AllNames(db, db.Organizations(), bson.M{})
	if err != nil {
		return
	}

	fields = []resource.FilterField{
		{
			Key:         "id",
			Label:       "Authority ID",
			Placeholder: "Authority ID",
		},
		{
			Key:         "name",
			Label:       "Name",
			Placeholder: "Name",
		},
		{
			Key:         "role",
			Label:       "Role",
			Placeholder: "Role",
		},
		{
			Key:         "principal",
			Label:       "Principal",
			Placeholder: "Principal",
		},
		{
			Key:     "organization",
			Label:   "Organization",
			Options: resource.SelectOptions(orgs),
		},
	}

	return
}

// NewEditor returns the settings form of a new authority.
func (p *Provider) NewEditor(db *database.Database) (
	editor resource.Editor, err error) {

	orgs, err := resource.AllNames(db, db.Organizations(), bson.M{})
	if err != nil {
		return
	}

	e := newEditor(&authority.Authority{
		Name:       "new-authority",
		Roles:      []string{},
		Principals: []string{},
	}, orgs)
	e.create = true
	editor = e

	return
}

// Load returns a page of authorities with the organization names
// resolved, the organizations are also the settings select options.
func (p *Provider) Load(db *database.Database, filter resource.Filter,
	page, pageCount int64) (result *resource.Page, err error) {

	query := buildQuery(filter)

	authrs, count, err := authority.GetAllPaged(db, &query, page, pageCount)
	if err != nil {
		return
	}

	orgs, err := resource.AllNames(db, db.Organizations(), bson.M{})
	if err != nil {
		return
	}
	orgNames := resource.NameMap(orgs)

	items := make([]resource.Item, 0, len(authrs))
	for _, authr := range authrs {
		items = append(items, &Item{
			authr: authr,
			orgs:  orgs,
			org:   orgNames[authr.Organization],
		})
	}

	result = &resource.Page{
		Items: items,
		Count: count,
	}

	return
}

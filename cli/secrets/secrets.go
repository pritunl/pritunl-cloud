// Package secrets provides the secrets tab, a paged list of the provider
// API secrets in the cluster from the admin perspective.
package secrets

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/secret"
)

const (
	// cardFields is the number of fields on each secret card.
	cardFields = 3

	// nodeSecret is the organization label of secrets without an
	// organization, these are available to the nodes.
	nodeSecret = "Node Secret"
)

type Provider struct{}

func New() *Provider {
	return &Provider{}
}

func (p *Provider) Title() string {
	return "Secrets"
}

func (p *Provider) Empty() string {
	return "No secrets"
}

func (p *Provider) Singular() string {
	return "secret"
}

func (p *Provider) FieldCount() int {
	return cardFields
}

// FilterFields mirrors the secrets filter of the web interface.
func (p *Provider) FilterFields(db *database.Database) (
	fields []resource.FilterField, err error) {

	fields = []resource.FilterField{
		{
			Key:         "id",
			Label:       "Secret ID",
			Placeholder: "Secret ID",
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

// NewEditor returns the settings form of a new secret.
func (p *Provider) NewEditor(db *database.Database) (
	editor resource.Editor, err error) {

	orgs, err := resource.AllNames(db, db.Organizations(), bson.M{})
	if err != nil {
		return
	}

	e := newEditor(&secret.Secret{
		Name: "new-secret",
		Type: secret.AWS,
	}, orgs)
	e.create = true
	editor = e

	return
}

// Load returns a page of secrets with the organization names resolved,
// the organizations are also the settings select options.
func (p *Provider) Load(db *database.Database, filter resource.Filter,
	page, pageCount int64) (result *resource.Page, err error) {

	query := buildQuery(filter)

	secrs, count, err := secret.GetAllPaged(db, &query, page, pageCount)
	if err != nil {
		return
	}

	orgs, err := resource.AllNames(db, db.Organizations(), bson.M{})
	if err != nil {
		return
	}
	orgNames := resource.NameMap(orgs)

	items := make([]resource.Item, 0, len(secrs))
	for _, secr := range secrs {
		items = append(items, &Item{
			secr: secr,
			orgs: orgs,
			org:  orgNames[secr.Organization],
		})
	}

	result = &resource.Page{
		Items: items,
		Count: count,
	}

	return
}

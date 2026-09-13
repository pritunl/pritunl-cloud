// Package storages provides the storages tab, a paged list of the object
// storages holding images in the cluster from the admin perspective.
package storages

import (
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/storage"
)

const (
	// cardFields is the number of fields on each storage card.
	cardFields = 3
)

type Provider struct{}

func New() *Provider {
	return &Provider{}
}

func (p *Provider) Title() string {
	return "Storages"
}

func (p *Provider) Empty() string {
	return "No storages"
}

func (p *Provider) Singular() string {
	return "storage"
}

func (p *Provider) FieldCount() int {
	return cardFields
}

// FilterFields mirrors the storages filter of the web interface.
func (p *Provider) FilterFields(db *database.Database) (
	fields []resource.FilterField, err error) {

	fields = []resource.FilterField{
		{
			Key:         "id",
			Label:       "Storage ID",
			Placeholder: "Storage ID",
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

// NewEditor returns the settings form of a new storage.
func (p *Provider) NewEditor(db *database.Database) (
	editor resource.Editor, err error) {

	e := newEditor(&storage.Storage{
		Name: "new-storage",
		Type: storage.Public,
	})
	e.create = true
	editor = e

	return
}

// Load returns a page of storages.
func (p *Provider) Load(db *database.Database, filter resource.Filter,
	page, pageCount int64) (result *resource.Page, err error) {

	query := buildQuery(filter)

	stores, count, err := storage.GetAllPaged(db, &query, page, pageCount)
	if err != nil {
		return
	}

	items := make([]resource.Item, 0, len(stores))
	for _, store := range stores {
		items = append(items, &Item{
			store: store,
		})
	}

	result = &resource.Page{
		Items: items,
		Count: count,
	}

	return
}

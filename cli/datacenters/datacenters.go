// Package datacenters provides the datacenters tab, a paged list of the
// datacenters in the cluster from the admin perspective.
package datacenters

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/datacenter"
	"github.com/pritunl/pritunl-cloud/storage"
)

const (
	// cardFields is the number of fields on each datacenter card.
	cardFields = 3
)

type Provider struct{}

func New() *Provider {
	return &Provider{}
}

func (p *Provider) Title() string {
	return "Datacenters"
}

func (p *Provider) Empty() string {
	return "No datacenters"
}

func (p *Provider) Singular() string {
	return "datacenter"
}

func (p *Provider) FieldCount() int {
	return cardFields
}

// FilterFields mirrors the datacenters filter of the web interface.
func (p *Provider) FilterFields(db *database.Database) (
	fields []resource.FilterField, err error) {

	fields = []resource.FilterField{
		{
			Key:         "id",
			Label:       "Datacenter ID",
			Placeholder: "Datacenter ID",
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

// names are the select options of the datacenter settings.
type names struct {
	orgs     []*database.Named
	storages []*storage.Storage
}

func loadNames(db *database.Database) (nms *names, err error) {
	nms = &names{}

	nms.orgs, err = resource.AllNames(db, db.Organizations(), bson.M{})
	if err != nil {
		return
	}

	nms.storages, err = storage.GetAll(db)
	if err != nil {
		return
	}

	return
}

// NewEditor returns the settings form of a new datacenter.
func (p *Provider) NewEditor(db *database.Database) (
	editor resource.Editor, err error) {

	nms, err := loadNames(db)
	if err != nil {
		return
	}

	e := newEditor(&datacenter.Datacenter{
		Name: "new-datacenter",
	}, nms)
	e.create = true
	editor = e

	return
}

// Load returns a page of datacenters, the organizations and storages
// are the settings select options.
func (p *Provider) Load(db *database.Database, filter resource.Filter,
	page, pageCount int64) (result *resource.Page, err error) {

	query := buildQuery(filter)

	dcs, count, err := datacenter.GetAllPaged(db, &query, page, pageCount)
	if err != nil {
		return
	}

	nms, err := loadNames(db)
	if err != nil {
		return
	}

	items := make([]resource.Item, 0, len(dcs))
	for _, dc := range dcs {
		items = append(items, &Item{
			dc:    dc,
			names: nms,
		})
	}

	result = &resource.Page{
		Items: items,
		Count: count,
	}

	return
}

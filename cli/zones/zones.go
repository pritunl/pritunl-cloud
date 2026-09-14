// Package zones provides the zones tab, a paged list of the zones in the
// cluster from the admin perspective.
package zones

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/zone"
)

const (
	// cardFields is the number of fields on each zone card.
	cardFields = 2
)

type Provider struct{}

func New() *Provider {
	return &Provider{}
}

func (p *Provider) Title() string {
	return "Zones"
}

func (p *Provider) Empty() string {
	return "No zones"
}

func (p *Provider) Singular() string {
	return "zone"
}

func (p *Provider) FieldCount() int {
	return cardFields
}

// FilterFields mirrors the zones filter of the web interface.
func (p *Provider) FilterFields(db *database.Database) (
	fields []resource.FilterField, err error) {

	fields = []resource.FilterField{
		{
			Key:         "id",
			Label:       "Zone ID",
			Placeholder: "Zone ID",
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

// NewEditor returns the settings form of a new zone.
func (p *Provider) NewEditor(db *database.Database) (
	editor resource.Editor, err error) {

	dcs, err := resource.AllNames(db, db.Datacenters(), bson.M{})
	if err != nil {
		return
	}

	e := newEditor(&zone.Zone{
		Name: "new-zone",
	}, dcs)
	e.create = true
	editor = e

	return
}

// Load returns a page of zones with the datacenter names resolved.
func (p *Provider) Load(db *database.Database, filter resource.Filter,
	page, pageCount int64) (result *resource.Page, err error) {

	query := buildQuery(filter)

	zones, count, err := zone.GetAllPaged(db, &query, page, pageCount)
	if err != nil {
		return
	}

	dcIds := resource.IdSet{}
	for _, zne := range zones {
		dcIds.Add(zne.Datacenter)
	}

	dcNames, err := resource.Names(db, db.Datacenters(), dcIds.List())
	if err != nil {
		return
	}

	items := make([]resource.Item, 0, len(zones))
	for _, zne := range zones {
		items = append(items, &Item{
			zne:        zne,
			datacenter: dcNames[zne.Datacenter],
		})
	}

	result = &resource.Page{
		Items: items,
		Count: count,
	}

	return
}

// Package pools provides the disk pools tab, a paged list of the LVM
// volume group pools in the cluster from the admin perspective.
package pools

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/pool"
	"github.com/pritunl/pritunl-cloud/zone"
)

const (
	// cardFields is the number of fields on each pool card.
	cardFields = 3
)

type Provider struct{}

func New() *Provider {
	return &Provider{}
}

func (p *Provider) Title() string {
	return "Disk Pools"
}

func (p *Provider) Empty() string {
	return "No pools"
}

func (p *Provider) Singular() string {
	return "pool"
}

func (p *Provider) FieldCount() int {
	return cardFields
}

// FilterFields mirrors the pools filter of the web interface.
func (p *Provider) FilterFields(db *database.Database) (
	fields []resource.FilterField, err error) {

	fields = []resource.FilterField{
		{
			Key:         "id",
			Label:       "Pool ID",
			Placeholder: "Pool ID",
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
		{
			Key:         "vg_name",
			Label:       "Volume Group Name",
			Placeholder: "Volume Group Name",
		},
	}

	return
}

// names are the select options of the pool settings.
type names struct {
	datacenters []*database.Named
	zones       []*zone.Zone
}

func loadNames(db *database.Database) (nms *names, err error) {
	nms = &names{}

	nms.datacenters, err = resource.AllNames(db, db.Datacenters(), bson.M{})
	if err != nil {
		return
	}

	nms.zones, err = zone.GetAll(db)
	if err != nil {
		return
	}

	return
}

// NewEditor returns the settings form of a new pool.
func (p *Provider) NewEditor(db *database.Database) (
	editor resource.Editor, err error) {

	nms, err := loadNames(db)
	if err != nil {
		return
	}

	e := newEditor(&pool.Pool{
		Name: "new-pool",
		Type: pool.Lvm,
	}, nms)
	e.create = true
	editor = e

	return
}

// Load returns a page of pools with the zone and datacenter names
// resolved, the names are also the settings select options.
func (p *Provider) Load(db *database.Database, filter resource.Filter,
	page, pageCount int64) (result *resource.Page, err error) {

	query := buildQuery(filter)

	pools, count, err := pool.GetAllPaged(db, &query, page, pageCount)
	if err != nil {
		return
	}

	nms, err := loadNames(db)
	if err != nil {
		return
	}
	dcNames := resource.NameMap(nms.datacenters)

	zoneNames := map[bson.ObjectID]string{}
	for _, zne := range nms.zones {
		zoneNames[zne.Id] = zne.Name
	}

	items := make([]resource.Item, 0, len(pools))
	for _, pl := range pools {
		items = append(items, &Item{
			pl:         pl,
			names:      nms,
			zone:       zoneNames[pl.Zone],
			datacenter: dcNames[pl.Datacenter],
		})
	}

	result = &resource.Page{
		Items: items,
		Count: count,
	}

	return
}

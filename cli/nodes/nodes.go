// Package nodes provides the nodes tab, a paged list of the nodes in the
// cluster from the admin perspective.
package nodes

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/certificate"
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/cli/widget"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/node"
)

const (
	// cardFields is the number of fields on each node card.
	cardFields = 6
)

var typeFilterOptions = []widget.SelectOption{
	{Label: "Yes", Value: "true"},
	{Label: "No", Value: "false"},
}

// names are the select options of the node settings loaded with each
// page.
type names struct {
	zones        []*database.Named
	blocks       []*database.Named
	certificates []*database.Named
	zoneNames    map[bson.ObjectID]string
}

type Provider struct{}

func New() *Provider {
	return &Provider{}
}

func (p *Provider) Title() string {
	return "Nodes"
}

func (p *Provider) Empty() string {
	return "No nodes"
}

func (p *Provider) Singular() string {
	return "node"
}

func (p *Provider) FieldCount() int {
	return cardFields
}

// FilterFields mirrors the nodes filter of the web interface, the type
// checkboxes become yes, no or any selects.
func (p *Provider) FilterFields(db *database.Database) (
	fields []resource.FilterField, err error) {

	zones, err := resource.AllNames(db, db.Zones(), bson.M{})
	if err != nil {
		return
	}

	fields = []resource.FilterField{
		{
			Key:         "id",
			Label:       "Node ID",
			Placeholder: "Node ID",
		},
		{
			Key:         "name",
			Label:       "Name",
			Placeholder: "Name",
		},
		{
			Key:         "role",
			Label:       "Network Role",
			Placeholder: "Network Role",
		},
		{
			Key:     "zone",
			Label:   "Zone",
			Options: resource.SelectOptions(zones),
		},
		{
			Key:     node.Admin,
			Label:   "Admin",
			Options: typeFilterOptions,
		},
		{
			Key:     node.User,
			Label:   "User",
			Options: typeFilterOptions,
		},
		{
			Key:     node.Hypervisor,
			Label:   "Hypervisor",
			Options: typeFilterOptions,
		},
	}

	return
}

func loadNames(db *database.Database) (nms *names, err error) {
	nms = &names{
		zoneNames: map[bson.ObjectID]string{},
	}

	nms.zones, err = resource.AllNames(db, db.Zones(), bson.M{})
	if err != nil {
		return
	}
	for _, zne := range nms.zones {
		nms.zoneNames[zne.Id] = zne.Name
	}

	nms.blocks, err = resource.AllNames(db, db.Blocks(), bson.M{})
	if err != nil {
		return
	}

	nms.certificates, err = certificate.GetAllNames(db, &bson.M{})
	if err != nil {
		return
	}

	return
}

// Load returns a page of nodes with the zone names resolved and the
// settings select options.
func (p *Provider) Load(db *database.Database, filter resource.Filter,
	page, pageCount int64) (result *resource.Page, err error) {

	query := buildQuery(filter)

	ndes, count, err := node.GetAllPaged(db, &query, page, pageCount)
	if err != nil {
		return
	}

	nms, err := loadNames(db)
	if err != nil {
		return
	}

	items := make([]resource.Item, 0, len(ndes))
	for _, nde := range ndes {
		nde.Json()

		items = append(items, &Item{
			nde:   nde,
			names: nms,
			zone:  nms.zoneNames[nde.Zone],
		})
	}

	result = &resource.Page{
		Items: items,
		Count: count,
	}

	return
}

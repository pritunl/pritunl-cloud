// Package shapes provides the shapes tab, a paged list of the instance
// and disk shapes in the cluster from the admin perspective.
package shapes

import (
	"fmt"
	"sort"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/aggregate"
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/cli/widget"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/pool"
	"github.com/pritunl/pritunl-cloud/shape"
	"github.com/pritunl/pritunl-cloud/zone"
)

const (
	// cardFields is the number of fields on each shape card.
	cardFields = 4
)

type Provider struct{}

func New() *Provider {
	return &Provider{}
}

func (p *Provider) Title() string {
	return "Shapes"
}

func (p *Provider) Empty() string {
	return "No shapes"
}

func (p *Provider) Singular() string {
	return "shape"
}

func (p *Provider) FieldCount() int {
	return cardFields
}

// FilterFields mirrors the shapes filter of the web interface.
func (p *Provider) FilterFields(db *database.Database) (
	fields []resource.FilterField, err error) {

	fields = []resource.FilterField{
		{
			Key:         "id",
			Label:       "Shape ID",
			Placeholder: "Shape ID",
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

// loadPoolOptions returns the disk pool choices of the shape settings
// labeled with the zone of each pool, a shape can hold pools of
// several zones.
func loadPoolOptions(db *database.Database) (
	opts []widget.SelectOption, err error) {

	opts = []widget.SelectOption{}

	pools, err := pool.GetAll(db, &bson.M{})
	if err != nil {
		return
	}

	zones, err := zone.GetAll(db)
	if err != nil {
		return
	}

	zoneNames := map[bson.ObjectID]string{}
	for _, zne := range zones {
		zoneNames[zne.Id] = zne.Name
	}

	sort.Slice(pools, func(i, j int) bool {
		return pools[i].Name < pools[j].Name
	})

	for _, pl := range pools {
		label := pl.Name
		if zoneName := zoneNames[pl.Zone]; zoneName != "" {
			label = fmt.Sprintf("%s (%s)", pl.Name, zoneName)
		}
		opts = append(opts, widget.SelectOption{
			Label: label,
			Value: pl.Id.Hex(),
		})
	}

	return
}

// NewEditor returns the settings form of a new shape with the web
// defaults.
func (p *Provider) NewEditor(db *database.Database) (
	editor resource.Editor, err error) {

	pools, err := loadPoolOptions(db)
	if err != nil {
		return
	}

	e := newEditor(&shape.Shape{
		Name:       "new-shape",
		Type:       shape.Instance,
		Memory:     1024,
		Processors: 1,
		Flexible:   true,
		Roles:      []string{},
	}, pools)
	e.create = true
	editor = e

	return
}

// Load returns a page of shapes with the node count of each shape and
// the pool choices of the settings form.
func (p *Provider) Load(db *database.Database, filter resource.Filter,
	page, pageCount int64) (result *resource.Page, err error) {

	query := buildQuery(filter)

	shapes, count, err := aggregate.GetShapePaged(db, &query, page,
		pageCount)
	if err != nil {
		return
	}

	pools, err := loadPoolOptions(db)
	if err != nil {
		return
	}

	items := make([]resource.Item, 0, len(shapes))
	for _, shpe := range shapes {
		items = append(items, &Item{
			shpe:  shpe,
			pools: pools,
		})
	}

	result = &resource.Page{
		Items: items,
		Count: count,
	}

	return
}

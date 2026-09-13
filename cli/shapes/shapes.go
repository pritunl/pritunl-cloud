// Package shapes provides the shapes tab, a paged list of the instance
// shapes in the cluster from the admin perspective.
package shapes

import (
	"github.com/pritunl/pritunl-cloud/aggregate"
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/shape"
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

// NewEditor returns the settings form of a new shape with the web
// defaults.
func (p *Provider) NewEditor(db *database.Database) (
	editor resource.Editor, err error) {

	e := newEditor(&shape.Shape{
		Name:       "new-shape",
		Memory:     1024,
		Processors: 1,
		Flexible:   true,
		Roles:      []string{},
	})
	e.create = true
	editor = e

	return
}

// Load returns a page of shapes with the node count of each shape.
func (p *Provider) Load(db *database.Database, filter resource.Filter,
	page, pageCount int64) (result *resource.Page, err error) {

	query := buildQuery(filter)

	shapes, count, err := aggregate.GetShapePaged(db, &query, page,
		pageCount)
	if err != nil {
		return
	}

	items := make([]resource.Item, 0, len(shapes))
	for _, shpe := range shapes {
		items = append(items, &Item{
			shpe: shpe,
		})
	}

	result = &resource.Page{
		Items: items,
		Count: count,
	}

	return
}

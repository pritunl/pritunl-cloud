// Package blocks provides the IP blocks tab, a paged list of the address
// blocks in the cluster from the admin perspective.
package blocks

import (
	"github.com/pritunl/pritunl-cloud/aggregate"
	"github.com/pritunl/pritunl-cloud/block"
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/database"
)

const (
	// cardFields is the number of fields on each block card.
	cardFields = 4
)

type Provider struct{}

func New() *Provider {
	return &Provider{}
}

func (p *Provider) Title() string {
	return "IP Blocks"
}

func (p *Provider) Empty() string {
	return "No blocks"
}

func (p *Provider) Singular() string {
	return "block"
}

func (p *Provider) FieldCount() int {
	return cardFields
}

// FilterFields mirrors the blocks filter of the web interface.
func (p *Provider) FilterFields(db *database.Database) (
	fields []resource.FilterField, err error) {

	fields = []resource.FilterField{
		{
			Key:         "id",
			Label:       "Block ID",
			Placeholder: "Block ID",
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

// NewEditor returns the settings form of a new block.
func (p *Provider) NewEditor(db *database.Database) (
	editor resource.Editor, err error) {

	e := newEditor(&aggregate.BlockAggregate{
		Block: block.Block{
			Name:     "new-block",
			Type:     block.IPv4,
			Subnets:  []string{},
			Subnets6: []string{},
			Excludes: []string{},
		},
	})
	e.create = true
	editor = e

	return
}

// Load returns a page of blocks with the available and total address
// counts.
func (p *Provider) Load(db *database.Database, filter resource.Filter,
	page, pageCount int64) (result *resource.Page, err error) {

	query := buildQuery(filter)

	blocks, count, err := aggregate.GetBlockPaged(db, &query, page,
		pageCount)
	if err != nil {
		return
	}

	items := make([]resource.Item, 0, len(blocks))
	for _, blck := range blocks {
		items = append(items, &Item{
			blck: blck,
		})
	}

	result = &resource.Page{
		Items: items,
		Count: count,
	}

	return
}

// Package images provides the images tab, a paged list of the disk
// images on the storages from the admin perspective. Images are created
// by snapshots and storage syncs so the tab has no create form.
package images

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/cli/widget"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/image"
)

const (
	// cardFields is the number of fields on each image card.
	cardFields = 4
)

type Provider struct{}

func New() *Provider {
	return &Provider{}
}

func (p *Provider) Title() string {
	return "Images"
}

func (p *Provider) Empty() string {
	return "No images"
}

func (p *Provider) Singular() string {
	return "image"
}

func (p *Provider) FieldCount() int {
	return cardFields
}

// FilterFields mirrors the images filter of the web interface.
func (p *Provider) FilterFields(db *database.Database) (
	fields []resource.FilterField, err error) {

	orgs, err := resource.AllNames(db, db.Organizations(), bson.M{})
	if err != nil {
		return
	}

	fields = []resource.FilterField{
		{
			Key:         "id",
			Label:       "Image ID",
			Placeholder: "Image ID",
		},
		{
			Key:         "name",
			Label:       "Name",
			Placeholder: "Name",
		},
		{
			Key:   "type",
			Label: "Type",
			Options: []widget.SelectOption{
				{Label: "Private", Value: "private"},
				{Label: "Public", Value: "public"},
			},
		},
		{
			Key:     "organization",
			Label:   "Organization",
			Options: resource.SelectOptions(orgs),
		},
	}

	return
}

// Load returns a page of images with the organization and storage names
// resolved.
func (p *Provider) Load(db *database.Database, filter resource.Filter,
	page, pageCount int64) (result *resource.Page, err error) {

	query := buildQuery(filter)

	images, count, err := image.GetAllPaged(db, &query, page, pageCount)
	if err != nil {
		return
	}

	orgIds := resource.IdSet{}
	storeIds := resource.IdSet{}
	for _, img := range images {
		orgIds.Add(img.Organization)
		storeIds.Add(img.Storage)
	}

	orgNames, err := resource.Names(db, db.Organizations(), orgIds.List())
	if err != nil {
		return
	}

	storeNames, err := resource.Names(db, db.Storages(), storeIds.List())
	if err != nil {
		return
	}

	items := make([]resource.Item, 0, len(images))
	for _, img := range images {
		img.Json()

		items = append(items, &Item{
			img:     img,
			org:     orgNames[img.Organization],
			storage: storeNames[img.Storage],
		})
	}

	result = &resource.Page{
		Items: items,
		Count: count,
	}

	return
}

// Package vpcs provides the VPCs tab, a paged list of the virtual private
// clouds in the cluster from the admin perspective.
package vpcs

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/vpc"
)

const (
	// cardFields is the number of fields on each VPC card.
	cardFields = 4
)

type Provider struct{}

func New() *Provider {
	return &Provider{}
}

func (p *Provider) Title() string {
	return "VPCs"
}

func (p *Provider) Empty() string {
	return "No VPCs"
}

func (p *Provider) Singular() string {
	return "VPC"
}

func (p *Provider) FieldCount() int {
	return cardFields
}

// names are the select options of the VPC settings.
type names struct {
	orgs        []*database.Named
	datacenters []*database.Named
}

func loadNames(db *database.Database) (nms *names, err error) {
	nms = &names{}

	nms.orgs, err = resource.AllNames(db, db.Organizations(), bson.M{})
	if err != nil {
		return
	}

	nms.datacenters, err = resource.AllNames(db, db.Datacenters(), bson.M{})
	if err != nil {
		return
	}

	return
}

// FilterFields mirrors the VPCs filter of the web interface.
func (p *Provider) FilterFields(db *database.Database) (
	fields []resource.FilterField, err error) {

	nms, err := loadNames(db)
	if err != nil {
		return
	}

	fields = []resource.FilterField{
		{
			Key:         "id",
			Label:       "VPC ID",
			Placeholder: "VPC ID",
		},
		{
			Key:         "name",
			Label:       "Name",
			Placeholder: "Name",
		},
		{
			Key:         "network",
			Label:       "Network",
			Placeholder: "Network",
		},
		{
			Key:     "datacenter",
			Label:   "Datacenter",
			Options: resource.SelectOptions(nms.datacenters),
		},
		{
			Key:     "organization",
			Label:   "Organization",
			Options: resource.SelectOptions(nms.orgs),
		},
	}

	return
}

// NewEditor returns the settings form of a new VPC, the organization
// and datacenter default to the first ones like the web interface.
func (p *Provider) NewEditor(db *database.Database) (
	editor resource.Editor, err error) {

	nms, err := loadNames(db)
	if err != nil {
		return
	}

	vc := &vpc.Vpc{
		Name:    "new-vpc",
		Subnets: []*vpc.Subnet{},
		Routes:  []*vpc.Route{},
		Maps:    []*vpc.Map{},
		Arps:    []*vpc.Arp{},
	}
	if len(nms.orgs) > 0 {
		vc.Organization = nms.orgs[0].Id
	}
	if len(nms.datacenters) > 0 {
		vc.Datacenter = nms.datacenters[0].Id
	}

	e := newEditor(vc, nms, true)
	editor = e

	return
}

// Load returns a page of VPCs with the organization and datacenter names
// resolved.
func (p *Provider) Load(db *database.Database, filter resource.Filter,
	page, pageCount int64) (result *resource.Page, err error) {

	query := buildQuery(filter)

	vpcs, count, err := vpc.GetAllPaged(db, &query, page, pageCount)
	if err != nil {
		return
	}

	nms, err := loadNames(db)
	if err != nil {
		return
	}
	orgNames := resource.NameMap(nms.orgs)
	dcNames := resource.NameMap(nms.datacenters)

	items := make([]resource.Item, 0, len(vpcs))
	for _, vc := range vpcs {
		vc.Json()

		items = append(items, &Item{
			vc:         vc,
			names:      nms,
			org:        orgNames[vc.Organization],
			datacenter: dcNames[vc.Datacenter],
		})
	}

	result = &resource.Page{
		Items: items,
		Count: count,
	}

	return
}

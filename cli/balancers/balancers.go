// Package balancers provides the load balancers tab, a paged list of the
// HTTP load balancers in the cluster from the admin perspective.
package balancers

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/balancer"
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/database"
)

const (
	// cardFields is the number of fields on each balancer card.
	cardFields = 4
)

type Provider struct{}

func New() *Provider {
	return &Provider{}
}

func (p *Provider) Title() string {
	return "Load Balancers"
}

func (p *Provider) Empty() string {
	return "No load balancers"
}

func (p *Provider) Singular() string {
	return "balancer"
}

func (p *Provider) FieldCount() int {
	return cardFields
}

// FilterFields mirrors the balancers filter of the web interface.
func (p *Provider) FilterFields(db *database.Database) (
	fields []resource.FilterField, err error) {

	orgs, err := resource.AllNames(db, db.Organizations(), bson.M{})
	if err != nil {
		return
	}

	fields = []resource.FilterField{
		{
			Key:         "id",
			Label:       "Balancer ID",
			Placeholder: "Balancer ID",
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
			Key:     "organization",
			Label:   "Organization",
			Options: resource.SelectOptions(orgs),
		},
	}

	return
}

// names are the select options of the balancer settings.
type names struct {
	orgs         []*database.Named
	datacenters  []*database.Named
	certificates []*resource.OrgNamed
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

	nms.certificates, err = resource.AllOrgNames(db, db.Certificates(),
		bson.M{})
	if err != nil {
		return
	}

	return
}

// NewEditor returns the settings form of a new balancer.
func (p *Provider) NewEditor(db *database.Database) (
	editor resource.Editor, err error) {

	nms, err := loadNames(db)
	if err != nil {
		return
	}

	e := newEditor(&balancer.Balancer{
		Name:         "new-balancer",
		Type:         balancer.Http,
		Certificates: []bson.ObjectID{},
		Domains:      []*balancer.Domain{},
		Backends:     []*balancer.Backend{},
	}, nms)
	e.create = true
	editor = e

	return
}

// Load returns a page of balancers with the organization and datacenter
// names resolved, the names are also the settings select options.
func (p *Provider) Load(db *database.Database, filter resource.Filter,
	page, pageCount int64) (result *resource.Page, err error) {

	query := buildQuery(filter)

	balncs, count, err := balancer.GetAllPaged(db, &query, page, pageCount)
	if err != nil {
		return
	}

	nms, err := loadNames(db)
	if err != nil {
		return
	}
	orgNames := resource.NameMap(nms.orgs)
	dcNames := resource.NameMap(nms.datacenters)

	items := make([]resource.Item, 0, len(balncs))
	for _, balnc := range balncs {
		balnc.Json()

		items = append(items, &Item{
			balnc:      balnc,
			names:      nms,
			org:        orgNames[balnc.Organization],
			datacenter: dcNames[balnc.Datacenter],
		})
	}

	result = &resource.Page{
		Items: items,
		Count: count,
	}

	return
}

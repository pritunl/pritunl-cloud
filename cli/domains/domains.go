// Package domains provides the domains tab, a paged list of the DNS
// domains in the cluster from the admin perspective.
package domains

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/aggregate"
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/domain"
)

const (
	// cardFields is the number of fields on each domain card.
	cardFields = 4

	// nodeDomain is the organization label of domains without an
	// organization.
	nodeDomain = "Node Domain"
)

type Provider struct{}

func New() *Provider {
	return &Provider{}
}

func (p *Provider) Title() string {
	return "Domains"
}

func (p *Provider) Empty() string {
	return "No domains"
}

func (p *Provider) Singular() string {
	return "domain"
}

func (p *Provider) FieldCount() int {
	return cardFields
}

// FilterFields mirrors the domains filter of the web interface.
func (p *Provider) FilterFields(db *database.Database) (
	fields []resource.FilterField, err error) {

	orgs, err := resource.AllNames(db, db.Organizations(), bson.M{})
	if err != nil {
		return
	}

	fields = []resource.FilterField{
		{
			Key:         "id",
			Label:       "Domain ID",
			Placeholder: "Domain ID",
		},
		{
			Key:         "name",
			Label:       "Name",
			Placeholder: "Name",
		},
		{
			Key:     "organization",
			Label:   "Organization",
			Options: resource.SelectOptions(orgs),
		},
	}

	return
}

// names are the select options of the domain settings.
type names struct {
	orgs    []*database.Named
	secrets []*resource.OrgNamed
}

func loadNames(db *database.Database) (nms *names, err error) {
	nms = &names{}

	nms.orgs, err = resource.AllNames(db, db.Organizations(), bson.M{})
	if err != nil {
		return
	}

	nms.secrets, err = resource.AllOrgNames(db, db.Secrets(), bson.M{})
	if err != nil {
		return
	}

	return
}

// NewEditor returns the settings form of a new domain, the name default
// is the one of the admin create handler and the organization the first
// organization like the web interface.
func (p *Provider) NewEditor(db *database.Database) (
	editor resource.Editor, err error) {

	nms, err := loadNames(db)
	if err != nil {
		return
	}

	domn := &domain.Domain{
		Name:    "new.domain",
		Type:    domain.Local,
		Records: []*domain.Record{},
	}
	if len(nms.orgs) > 0 {
		domn.Organization = nms.orgs[0].Id
	}

	e := newEditor(domn, nms)
	e.create = true
	editor = e

	return
}

// Load returns a page of domains with their records and the
// organization names resolved, the organizations and secrets are the
// settings select options.
func (p *Provider) Load(db *database.Database, filter resource.Filter,
	page, pageCount int64) (result *resource.Page, err error) {

	query := buildQuery(filter)

	domns, count, err := aggregate.GetDomainPaged(db, &query, page,
		pageCount)
	if err != nil {
		return
	}

	nms, err := loadNames(db)
	if err != nil {
		return
	}
	orgNames := resource.NameMap(nms.orgs)

	items := make([]resource.Item, 0, len(domns))
	for _, domn := range domns {
		items = append(items, &Item{
			domn:  domn,
			names: nms,
			org:   orgNames[domn.Organization],
		})
	}

	result = &resource.Page{
		Items: items,
		Count: count,
	}

	return
}

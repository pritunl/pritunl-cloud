// Package plans provides the plans tab, a paged list of the pod
// scheduling plans in the cluster from the admin perspective.
package plans

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/plan"
)

const (
	// cardFields is the number of fields on each plan card.
	cardFields = 3
)

type Provider struct{}

func New() *Provider {
	return &Provider{}
}

func (p *Provider) Title() string {
	return "Plans"
}

func (p *Provider) Empty() string {
	return "No plans"
}

func (p *Provider) Singular() string {
	return "plan"
}

func (p *Provider) FieldCount() int {
	return cardFields
}

// FilterFields mirrors the plans filter of the web interface.
func (p *Provider) FilterFields(db *database.Database) (
	fields []resource.FilterField, err error) {

	orgs, err := resource.AllNames(db, db.Organizations(), bson.M{})
	if err != nil {
		return
	}

	fields = []resource.FilterField{
		{
			Key:         "id",
			Label:       "Plan ID",
			Placeholder: "Plan ID",
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

// NewEditor returns the settings form of a new plan, the name default
// is the one of the admin create handler and the organization the first
// organization like the web interface.
func (p *Provider) NewEditor(db *database.Database) (
	editor resource.Editor, err error) {

	orgs, err := resource.AllNames(db, db.Organizations(), bson.M{})
	if err != nil {
		return
	}

	pln := &plan.Plan{
		Name:       "new-plan",
		Statements: []*plan.Statement{},
	}
	if len(orgs) > 0 {
		pln.Organization = orgs[0].Id
	}

	e := newEditor(pln, orgs)
	e.create = true
	editor = e

	return
}

// Load returns a page of plans with the organization names resolved,
// the organizations are also the settings select options.
func (p *Provider) Load(db *database.Database, filter resource.Filter,
	page, pageCount int64) (result *resource.Page, err error) {

	query := buildQuery(filter)

	plans, count, err := plan.GetAllPaged(db, &query, page, pageCount)
	if err != nil {
		return
	}

	orgs, err := resource.AllNames(db, db.Organizations(), bson.M{})
	if err != nil {
		return
	}
	orgNames := resource.NameMap(orgs)

	items := make([]resource.Item, 0, len(plans))
	for _, pln := range plans {
		items = append(items, &Item{
			pln:  pln,
			orgs: orgs,
			org:  orgNames[pln.Organization],
		})
	}

	result = &resource.Page{
		Items: items,
		Count: count,
	}

	return
}

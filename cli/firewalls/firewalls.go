// Package firewalls provides the firewalls tab, a paged list of the
// firewalls in the cluster from the admin perspective.
package firewalls

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/firewall"
)

const (
	// cardFields is the number of fields on each firewall card.
	cardFields = 4

	// nodeFirewall is the organization label of firewalls without an
	// organization, these apply to the nodes.
	nodeFirewall = "Node Firewall"
)

type Provider struct{}

func New() *Provider {
	return &Provider{}
}

func (p *Provider) Title() string {
	return "Firewalls"
}

func (p *Provider) Empty() string {
	return "No firewalls"
}

func (p *Provider) Singular() string {
	return "firewall"
}

func (p *Provider) FieldCount() int {
	return cardFields
}

// FilterFields mirrors the firewalls filter of the web interface.
func (p *Provider) FilterFields(db *database.Database) (
	fields []resource.FilterField, err error) {

	orgs, err := resource.AllNames(db, db.Organizations(), bson.M{})
	if err != nil {
		return
	}

	fields = []resource.FilterField{
		{
			Key:         "id",
			Label:       "Firewall ID",
			Placeholder: "Firewall ID",
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
			Key:         "role",
			Label:       "Network Role",
			Placeholder: "Network Role",
		},
		{
			Key:     "organization",
			Label:   "Organization",
			Options: resource.SelectOptions(orgs),
		},
	}

	return
}

// NewEditor returns the settings form of a new firewall.
func (p *Provider) NewEditor(db *database.Database) (
	editor resource.Editor, err error) {

	orgs, err := resource.AllNames(db, db.Organizations(), bson.M{})
	if err != nil {
		return
	}

	e := newEditor(newFirewall(), orgs)
	e.create = true
	editor = e

	return
}

// Load returns a page of firewalls with the organization names resolved,
// the organizations are also the settings select options.
func (p *Provider) Load(db *database.Database, filter resource.Filter,
	page, pageCount int64) (result *resource.Page, err error) {

	query := buildQuery(filter)

	fires, count, err := firewall.GetAllPaged(db, &query, page, pageCount)
	if err != nil {
		return
	}

	orgs, err := resource.AllNames(db, db.Organizations(), bson.M{})
	if err != nil {
		return
	}

	orgNames := map[bson.ObjectID]string{}
	for _, org := range orgs {
		orgNames[org.Id] = org.Name
	}

	items := make([]resource.Item, 0, len(fires))
	for _, fire := range fires {
		items = append(items, &Item{
			fire: fire,
			orgs: orgs,
			org:  orgNames[fire.Organization],
		})
	}

	result = &resource.Page{
		Items: items,
		Count: count,
	}

	return
}

// Package certificates provides the certificates tab, a paged list of
// the TLS certificates in the cluster from the admin perspective.
package certificates

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/certificate"
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/database"
)

const (
	// cardFields is the number of fields on each certificate card.
	cardFields = 4

	// nodeCertificate is the organization label of certificates without
	// an organization, these are available to the nodes.
	nodeCertificate = "Node Certificate"
)

type Provider struct{}

func New() *Provider {
	return &Provider{}
}

func (p *Provider) Title() string {
	return "Certificates"
}

func (p *Provider) Empty() string {
	return "No certificates"
}

func (p *Provider) Singular() string {
	return "certificate"
}

func (p *Provider) FieldCount() int {
	return cardFields
}

// FilterFields mirrors the certificates filter of the web interface.
func (p *Provider) FilterFields(db *database.Database) (
	fields []resource.FilterField, err error) {

	fields = []resource.FilterField{
		{
			Key:         "id",
			Label:       "Certificate ID",
			Placeholder: "Certificate ID",
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

// names are the select options of the certificate settings.
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

// NewEditor returns the settings form of a new certificate.
func (p *Provider) NewEditor(db *database.Database) (
	editor resource.Editor, err error) {

	nms, err := loadNames(db)
	if err != nil {
		return
	}

	e := newEditor(&certificate.Certificate{
		Name:        "new-certificate",
		Type:        certificate.Text,
		AcmeDomains: []string{},
	}, nms)
	e.create = true
	editor = e

	return
}

// Load returns a page of certificates with the organization names
// resolved, the organizations and secrets are the settings select
// options.
func (p *Provider) Load(db *database.Database, filter resource.Filter,
	page, pageCount int64) (result *resource.Page, err error) {

	query := buildQuery(filter)

	certs, count, err := certificate.GetAllPaged(db, &query, page,
		pageCount)
	if err != nil {
		return
	}

	nms, err := loadNames(db)
	if err != nil {
		return
	}
	orgNames := resource.NameMap(nms.orgs)

	items := make([]resource.Item, 0, len(certs))
	for _, cert := range certs {
		items = append(items, &Item{
			cert:  cert,
			names: nms,
			org:   orgNames[cert.Organization],
		})
	}

	result = &resource.Page{
		Items: items,
		Count: count,
	}

	return
}

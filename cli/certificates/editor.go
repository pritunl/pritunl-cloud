package certificates

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/acme"
	"github.com/pritunl/pritunl-cloud/certificate"
	"github.com/pritunl/pritunl-cloud/cli/form"
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/cli/widget"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/sirupsen/logrus"
)

var typeOptions = []widget.SelectOption{
	{Label: "Text", Value: certificate.Text},
	{Label: "LetsEncrypt", Value: certificate.LetsEncrypt},
}

var acmeTypeOptions = []widget.SelectOption{
	{Label: "HTTP", Value: certificate.AcmeHTTP},
	{Label: "DNS TXT", Value: certificate.AcmeDNS},
}

var acmeAuthOptions = []widget.SelectOption{
	{Label: "AWS", Value: certificate.AcmeAWS},
	{Label: "Cloudflare", Value: certificate.AcmeCloudflare},
	{Label: "Oracle Cloud", Value: certificate.AcmeOracleCloud},
	{Label: "Google Cloud", Value: certificate.AcmeGoogleCloud},
}

// editor is the certificate settings form mirroring the inputs of the
// detailed certificate view in the web interface.
type editor struct {
	cert   *certificate.Certificate
	names  *names
	create bool
	form   *form.Form

	name         *form.Text
	comment      *form.Area
	key          *form.Area
	certificate  *form.Area
	acmeDomains  *form.Rows
	typ          *form.Select
	organization *form.Select
	acmeType     *form.Select
	acmeAuth     *form.Select
	acmeSecret   *form.Select
}

// domainRow creates the domain input of a LetsEncrypt domain.
func domainRow(domain string) *form.Row {
	return &form.Row{
		Cells: []form.Input{
			form.NewCell("Domain", domain),
		},
		Tag: domain,
	}
}

func newDomainRow() *form.Row {
	row := domainRow("")
	row.Tag = nil
	return row
}

func newEditor(cert *certificate.Certificate, nms *names) *editor {
	e := &editor{
		cert:  cert,
		names: nms,
	}

	e.name = form.NewText("Name", "Name", cert.Name)
	e.comment = form.NewArea("Comment", "Certificate comment",
		cert.Comment, 4)

	// The key and chain are managed by LetsEncrypt for acme certificates
	textOnly := func() bool {
		return e.typ.Value() != certificate.Text
	}
	e.key = form.NewArea("Private Key", "Private key", cert.Key, 6)
	e.key.Lock = textOnly
	e.certificate = form.NewArea("Certificate Chain", "Certificate chain",
		cert.Certificate, 6)
	e.certificate.Lock = textOnly

	letsEncrypt := func() bool {
		return e.typ.Value() != certificate.LetsEncrypt
	}
	e.acmeDomains = form.NewRows("LetsEncrypt Domains", "No domains",
		"Add Domain", []string{"Domain"}, newDomainRow)
	for _, domain := range cert.AcmeDomains {
		e.acmeDomains.Add(domainRow(domain))
	}
	e.acmeDomains.Hide = letsEncrypt

	e.typ = form.NewSelect("Type", typeOptions, cert.Type)

	orgOpts := resource.WithSelect(nodeCertificate,
		resource.SelectOptions(nms.orgs))
	e.organization = form.NewSelect("Organization", orgOpts,
		resource.IdHex(cert.Organization))

	e.acmeType = form.NewSelect("LetsEncrypt Verification Type",
		acmeTypeOptions, cert.AcmeType)
	e.acmeType.Hide = letsEncrypt

	acmeDns := func() bool {
		return letsEncrypt() || e.acmeType.Value() != certificate.AcmeDNS
	}
	e.acmeAuth = form.NewSelect("LetsEncrypt Verification Provider",
		acmeAuthOptions, cert.AcmeAuth)
	e.acmeAuth.Hide = acmeDns

	e.acmeSecret = form.NewSelect("LetsEncrypt Verification Secret", nil,
		resource.IdHex(cert.AcmeSecret))
	e.acmeSecret.Dynamic = e.secretOptions
	e.acmeSecret.Hide = acmeDns

	e.form = form.New(
		e.name,
		e.comment,
		e.key,
		e.certificate,
		e.acmeDomains,
		e.typ,
		e.organization,
		e.acmeType,
		e.acmeAuth,
		e.acmeSecret,
	)

	return e
}

// secretOptions lists the secrets of the selected organization like the
// web interface.
func (e *editor) secretOptions() []widget.SelectOption {
	opts := resource.OrgSelectOptions(e.names.secrets,
		resource.ParseId(e.organization.Value()))
	if len(opts) == 0 {
		return resource.WithSelect("No Secrets", nil)
	}
	return resource.WithSelect("Select Secret", opts)
}

func (e *editor) Form() *form.Form {
	return e.form
}

func (e *editor) buildDomains() []string {
	domains := []string{}
	for _, row := range e.acmeDomains.Rows() {
		domains = append(domains, row.Cells[0].(*form.Text).Value())
	}
	return domains
}

// apply sets the form values on the certificate, the key and chain are
// only written for text certificates like the admin handler.
func (e *editor) apply(cert *certificate.Certificate, fields set.Set) {
	cert.Name = e.name.Value()
	cert.Comment = e.comment.Value()
	cert.Organization = resource.ParseId(e.organization.Value())
	cert.Type = e.typ.Value()
	cert.AcmeDomains = e.buildDomains()
	cert.AcmeType = e.acmeType.Value()
	cert.AcmeAuth = e.acmeAuth.Value()
	cert.AcmeSecret = resource.ParseId(e.acmeSecret.Value())

	if cert.Type != certificate.LetsEncrypt {
		cert.Key = e.key.Value()
		cert.Certificate = e.certificate.Value()
		if fields != nil {
			fields.Add("key")
			fields.Add("certificate")
		}
	}
}

// Save applies the form to the current certificate the same way as the
// admin certificate handler, or inserts a new certificate. LetsEncrypt
// certificates are renewed in the background after saving.
func (e *editor) Save(db *database.Database) (err error) {
	fields := set.NewSet(
		"name",
		"comment",
		"organization",
		"type",
		"acme_domains",
		"acme_type",
		"acme_auth",
		"acme_secret",
		"info",
	)

	var cert *certificate.Certificate
	if e.create {
		cert = &certificate.Certificate{}
	} else {
		cert, err = certificate.Get(db, e.cert.Id)
		if err != nil {
			return
		}
	}

	e.apply(cert, fields)

	errData, err := cert.Validate(db)
	if err != nil {
		return
	}
	if errData != nil {
		err = &errortypes.ParseError{
			errors.New("certificates: " + errData.Message),
		}
		return
	}

	if e.create {
		err = cert.Insert(db)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"certificate_id": cert.Id.Hex(),
		}).Info("tui: Certificate created")
	} else {
		err = cert.CommitFields(db, fields)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"certificate_id": cert.Id.Hex(),
		}).Info("tui: Certificate settings saved")
	}

	if cert.Type == certificate.LetsEncrypt {
		acme.RenewBackground(cert, false)
	}

	err = event.PublishDispatch(db, "certificate.change")
	if err != nil {
		return
	}

	return
}

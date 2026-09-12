package certificates

import (
	"strings"

	"github.com/pritunl/pritunl-cloud/acme"
	"github.com/pritunl/pritunl-cloud/certificate"
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/cli/widget"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/sirupsen/logrus"
)

// Item is a certificate card with the organization name resolved.
type Item struct {
	cert  *certificate.Certificate
	names *names
	org   string
}

func (i *Item) Certificate() *certificate.Certificate {
	return i.cert
}

func (i *Item) Id() string {
	return i.cert.Id.Hex()
}

func (i *Item) Name() string {
	return i.cert.Name
}

func (i *Item) Tag() string {
	return i.cert.Id.Hex()
}

// organization returns the organization name or the node certificate
// label for certificates without one.
func (i *Item) organization() string {
	if i.cert.Organization.IsZero() {
		return nodeCertificate
	}
	return widget.Default(i.org, i.cert.Organization.Hex())
}

func typeLabel(typ string) string {
	for _, opt := range typeOptions {
		if opt.Value == typ {
			return opt.Label
		}
	}
	return typ
}

// info returns the parsed certificate info or an empty info.
func (i *Item) info() *certificate.Info {
	if i.cert.Info == nil {
		return &certificate.Info{}
	}
	return i.cert.Info
}

func (i *Item) Fields() []resource.Field {
	info := i.info()

	return []resource.Field{
		{
			Label: "Organization",
			Value: i.organization(),
		},
		{
			Label: "Type",
			Value: typeLabel(i.cert.Type),
		},
		{
			Label: "DNS Names",
			Value: strings.Join(info.DnsNames, ", "),
		},
		{
			Label: "Expires On",
			Value: resource.FormatTime(info.ExpiresOn),
		},
	}
}

// Info mirrors the fields of the detailed certificate view.
func (i *Item) Info() []widget.InfoField {
	info := i.info()

	return []widget.InfoField{
		{"ID", i.cert.Id.Hex()},
		{"Signature Algorithm", widget.Default(info.SignatureAlg, "Unknown")},
		{"Public Key Algorithm", widget.Default(info.PublicKeyAlg, "Unknown")},
		{"Issuer", widget.Default(info.Issuer, "Unknown")},
		{"Issued On", widget.Default(resource.FormatTime(info.IssuedOn),
			"Unknown")},
		{"Expires On", widget.Default(resource.FormatTime(info.ExpiresOn),
			"Unknown")},
		{"DNS Names", widget.JoinDefault(info.DnsNames, ", ", "Unknown")},
	}
}

// renew starts the LetsEncrypt renewal of the certificate the same way
// as the renew button of the web interface, which commits the
// certificate with the refresh flag.
func renew(certId string) func(db *database.Database) error {
	return func(db *database.Database) (err error) {
		cert, err := certificate.Get(db, resource.ParseId(certId))
		if err != nil {
			return
		}

		if cert.Type == certificate.LetsEncrypt {
			acme.RenewBackground(cert, true)
		}

		logrus.WithFields(logrus.Fields{
			"certificate_id": cert.Id.Hex(),
		}).Info("tui: Certificate renew started")

		err = event.PublishDispatch(db, "certificate.change")
		if err != nil {
			return
		}

		return
	}
}

// Actions mirrors the renew and delete buttons of the detailed
// certificate view, renew is shown for LetsEncrypt certificates.
func (i *Item) Actions() []resource.Action {
	actions := []resource.Action{}

	if i.cert.Type == certificate.LetsEncrypt {
		actions = append(actions, resource.Action{
			Key:     "w",
			Label:   "Renew",
			Status:  "Renewing",
			Confirm: "Renew the certificate " + i.cert.Name + "?",
			Run:     renew(i.cert.Id.Hex()),
		})
	}

	actions = append(actions, resource.DeleteAction("certificate",
		i.cert.Name, resource.DeleteRelated("certificate", "certificate",
			"certificate.change", i.cert.Id, certificate.Remove)))

	return actions
}

func (i *Item) Editor() resource.Editor {
	return newEditor(i.cert, i.names)
}

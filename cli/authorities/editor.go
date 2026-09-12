package authorities

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/authority"
	"github.com/pritunl/pritunl-cloud/cli/form"
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/cli/widget"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/sirupsen/logrus"
)

// commitFields are the fields written by the admin authority handler.
var commitFields = set.NewSet(
	"name",
	"comment",
	"type",
	"organization",
	"roles",
	"key",
	"principals",
	"certificate",
)

var typeOptions = []widget.SelectOption{
	{Label: "SSH Key", Value: authority.SshKey},
	{Label: "SSH Certificate", Value: authority.SshCertificate},
}

// editor is the authority settings form mirroring the inputs of the
// detailed authority view in the web interface.
type editor struct {
	authr  *authority.Authority
	create bool
	form   *form.Form

	name         *form.Text
	comment      *form.Area
	typ          *form.Select
	key          *form.Area
	certificate  *form.Area
	principals   *form.Tokens
	organization *form.Select
	roles        *form.Tokens
}

func newEditor(authr *authority.Authority, orgs []*database.Named) *editor {
	e := &editor{
		authr: authr,
	}

	e.name = form.NewText("Name", "Enter name", authr.Name)
	e.comment = form.NewArea("Comment", "Authority comment",
		authr.Comment, 4)

	e.typ = form.NewSelect("Type", typeOptions, authr.Type)

	e.key = form.NewArea("SSH Key", "Public key", authr.Key, 4)
	e.key.Hide = func() bool {
		return e.typ.Value() != authority.SshKey
	}

	e.certificate = form.NewArea("SSH Certificate", "Certificate authority",
		authr.Certificate, 4)
	e.certificate.Hide = func() bool {
		return e.typ.Value() != authority.SshCertificate
	}

	e.principals = form.NewTokens("Principals", "Add principal",
		authr.Principals)
	e.principals.Hide = e.certificate.Hide

	// The web interface selects the first organization when the
	// authority has none
	orgOpts := resource.SelectOptions(orgs)
	if len(orgOpts) == 0 {
		orgOpts = resource.WithSelect("No Organizations", nil)
	}
	e.organization = form.NewSelect("Organization", orgOpts,
		resource.IdHex(authr.Organization))
	e.organization.Lock = func() bool {
		return len(orgs) == 0
	}

	e.roles = form.NewTokens("Roles", "Add role", authr.Roles)

	e.form = form.New(
		e.name,
		e.comment,
		e.typ,
		e.key,
		e.certificate,
		e.principals,
		e.organization,
		e.roles,
	)

	return e
}

func (e *editor) Form() *form.Form {
	return e.form
}

// apply sets the form values on the authority.
func (e *editor) apply(authr *authority.Authority) {
	authr.Name = e.name.Value()
	authr.Comment = e.comment.Value()
	authr.Type = e.typ.Value()
	authr.Organization = resource.ParseId(e.organization.Value())
	authr.Roles = e.roles.Values()
	authr.Key = e.key.Value()
	authr.Principals = e.principals.Values()
	authr.Certificate = e.certificate.Value()
}

// Save applies the form to the current authority the same way as the
// admin authority handler, or inserts a new authority.
func (e *editor) Save(db *database.Database) (err error) {
	var authr *authority.Authority
	if e.create {
		authr = &authority.Authority{}
	} else {
		authr, err = authority.Get(db, e.authr.Id)
		if err != nil {
			return
		}
	}

	e.apply(authr)

	errData, err := authr.Validate(db)
	if err != nil {
		return
	}
	if errData != nil {
		err = &errortypes.ParseError{
			errors.New("authorities: " + errData.Message),
		}
		return
	}

	if e.create {
		err = authr.Insert(db)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"authority_id": authr.Id.Hex(),
		}).Info("tui: Authority created")
	} else {
		err = authr.CommitFields(db, commitFields)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"authority_id": authr.Id.Hex(),
		}).Info("tui: Authority settings saved")
	}

	err = event.PublishDispatch(db, "authority.change")
	if err != nil {
		return
	}

	return
}

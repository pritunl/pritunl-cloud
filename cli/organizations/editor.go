package organizations

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/cli/form"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/organization"
	"github.com/pritunl/pritunl-cloud/subscription"
	"github.com/sirupsen/logrus"
)

// commitFields are the fields written by the admin organization
// handler.
var commitFields = set.NewSet(
	"name",
	"comment",
	"roles",
)

// editor is the organization settings form mirroring the inputs of the
// detailed organization view in the web interface.
type editor struct {
	org    *organization.Organization
	create bool
	form   *form.Form

	name    *form.Text
	comment *form.Area
	roles   *form.Tokens
}

func newEditor(org *organization.Organization) *editor {
	e := &editor{
		org: org,
	}

	e.name = form.NewText("Name", "Name", org.Name)
	e.comment = form.NewArea("Comment", "Organization comment",
		org.Comment, 4)
	e.roles = form.NewTokens("Roles", "Add role", org.Roles)

	e.form = form.New(
		e.name,
		e.comment,
		e.roles,
	)

	return e
}

func (e *editor) Form() *form.Form {
	return e.form
}

func saveError(message string) error {
	return &errortypes.ParseError{
		errors.New("organizations: " + message),
	}
}

// Save applies the form to the current organization the same way as
// the admin organization handler, or inserts a new organization which
// requires a subscription once one exists.
func (e *editor) Save(db *database.Database) (err error) {
	var org *organization.Organization
	if e.create {
		if !subscription.Sub.Active {
			count, er := organization.Count(db)
			if er != nil {
				err = er
				return
			}

			if count > 0 {
				err = saveError(
					"Subscription required for multiple organizations")
				return
			}
		}

		org = &organization.Organization{}
	} else {
		org, err = organization.Get(db, e.org.Id)
		if err != nil {
			return
		}
	}

	org.Name = e.name.Value()
	org.Comment = e.comment.Value()
	org.Roles = e.roles.Values()

	errData, err := org.Validate(db)
	if err != nil {
		return
	}
	if errData != nil {
		err = saveError(errData.Message)
		return
	}

	if e.create {
		err = org.Insert(db)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"organization_id": org.Id.Hex(),
		}).Info("tui: Organization created")
	} else {
		err = org.CommitFields(db, commitFields)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"organization_id": org.Id.Hex(),
		}).Info("tui: Organization settings saved")
	}

	err = event.PublishDispatch(db, "organization.change")
	if err != nil {
		return
	}

	return
}

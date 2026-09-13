package plans

import (
	"strings"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/cli/form"
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/plan"
	"github.com/sirupsen/logrus"
)

// commitFields are the fields written by the admin plan handler.
var commitFields = set.NewSet(
	"name",
	"comment",
	"organization",
	"statements",
)

// editor is the plan settings form mirroring the inputs of the detailed
// plan view in the web interface, the statements are edited one per
// line like the web statement editor.
type editor struct {
	pln    *plan.Plan
	create bool
	form   *form.Form

	name         *form.Text
	comment      *form.Area
	organization *form.Select
	statements   *form.Area
}

func statementText(statements []*plan.Statement) string {
	lines := []string{}
	for _, statement := range statements {
		lines = append(lines, statement.Statement)
	}
	return strings.Join(lines, "\n")
}

func newEditor(pln *plan.Plan, orgs []*database.Named) *editor {
	e := &editor{
		pln: pln,
	}

	e.name = form.NewText("Name", "Enter name", pln.Name)
	e.comment = form.NewArea("Comment", "Plan comment", pln.Comment, 4)

	orgOpts := resource.SelectOptions(orgs)
	if len(orgOpts) == 0 {
		orgOpts = resource.WithSelect("No Organizations", nil)
	}
	e.organization = form.NewSelect("Organization", orgOpts,
		resource.IdHex(pln.Organization))

	e.statements = form.NewArea("Statements", "One statement per line",
		statementText(pln.Statements), 8)

	e.form = form.New(
		e.name,
		e.comment,
		e.organization,
		e.statements,
	)

	return e
}

func (e *editor) Form() *form.Form {
	return e.form
}

// buildStatements returns one statement per non empty line keeping the
// id of the current statement at the same position like the web
// statement editor, changed statements get a new id on update.
func (e *editor) buildStatements() []*plan.Statement {
	statements := []*plan.Statement{}
	for _, line := range strings.Split(e.statements.Value(), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		statement := &plan.Statement{
			Statement: line,
		}
		if i := len(statements); i < len(e.pln.Statements) {
			statement.Id = e.pln.Statements[i].Id
		}

		statements = append(statements, statement)
	}
	return statements
}

// apply sets the form values on the plan.
func (e *editor) apply(pln *plan.Plan) (err error) {
	pln.Name = e.name.Value()
	pln.Comment = e.comment.Value()
	pln.Organization = resource.ParseId(e.organization.Value())

	err = pln.UpdateStatements(e.buildStatements())
	if err != nil {
		return
	}

	return
}

// Save applies the form to the current plan the same way as the admin
// plan handler, or inserts a new plan.
func (e *editor) Save(db *database.Database) (err error) {
	var pln *plan.Plan
	if e.create {
		pln = &plan.Plan{}
	} else {
		pln, err = plan.Get(db, e.pln.Id)
		if err != nil {
			return
		}
	}

	err = e.apply(pln)
	if err != nil {
		return
	}

	errData, err := pln.Validate(db)
	if err != nil {
		return
	}
	if errData != nil {
		err = &errortypes.ParseError{
			errors.New("plans: " + errData.Message),
		}
		return
	}

	if e.create {
		err = pln.Insert(db)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"plan_id": pln.Id.Hex(),
		}).Info("tui: Plan created")
	} else {
		err = pln.CommitFields(db, commitFields)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"plan_id": pln.Id.Hex(),
		}).Info("tui: Plan settings saved")
	}

	err = event.PublishDispatch(db, "plan.change")
	if err != nil {
		return
	}

	return
}

package shapes

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/cli/form"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/shape"
	"github.com/sirupsen/logrus"
)

// commitFields are the fields written by the admin shape handler.
var commitFields = set.NewSet(
	"name",
	"type",
	"comment",
	"delete_protection",
	"roles",
	"flexible",
	"disk_type",
	"disk_pool",
	"memory",
	"processors",
)

// editor is the shape settings form mirroring the inputs of the
// detailed shape view in the web interface. The disk type and pool
// inputs are hidden in the web interface and keep their stored values.
type editor struct {
	shpe   *shape.Shape
	create bool
	form   *form.Form

	name             *form.Text
	comment          *form.Area
	flexible         *form.Toggle
	deleteProtection *form.Toggle
	memory           *form.Number
	processors       *form.Number
	roles            *form.Tokens
}

func newEditor(shpe *shape.Shape) *editor {
	e := &editor{
		shpe: shpe,
	}

	e.name = form.NewText("Name", "Enter name", shpe.Name)
	e.comment = form.NewArea("Comment", "Shape comment", shpe.Comment, 4)
	e.flexible = form.NewToggle("Flexible", shpe.Flexible)
	e.deleteProtection = form.NewToggle("Delete protection",
		shpe.DeleteProtection)
	e.memory = form.NewNumber("Memory Size (MB)", "Memory in megabytes",
		shpe.Memory, 256, 0, 1024)
	e.processors = form.NewNumber("Processors", "Number of processors",
		shpe.Processors, 1, 0, 1)
	e.roles = form.NewTokens("Roles", "Add role", shpe.Roles)

	e.form = form.New(
		e.name,
		e.comment,
		e.flexible,
		e.deleteProtection,
		e.memory,
		e.processors,
		e.roles,
	)

	return e
}

func (e *editor) Form() *form.Form {
	return e.form
}

// apply sets the form values on the shape.
func (e *editor) apply(shpe *shape.Shape) {
	shpe.Name = e.name.Value()
	shpe.Comment = e.comment.Value()
	shpe.DeleteProtection = e.deleteProtection.Value()
	shpe.Roles = e.roles.Values()
	shpe.Flexible = e.flexible.Value()
	shpe.Memory = e.memory.Int()
	shpe.Processors = e.processors.Int()
}

// Save applies the form to the current shape the same way as the admin
// shape handler, or inserts a new shape.
func (e *editor) Save(db *database.Database) (err error) {
	var shpe *shape.Shape
	if e.create {
		shpe = &shape.Shape{}
	} else {
		shpe, err = shape.Get(db, e.shpe.Id)
		if err != nil {
			return
		}
	}

	e.apply(shpe)

	errData, err := shpe.Validate(db)
	if err != nil {
		return
	}
	if errData != nil {
		err = &errortypes.ParseError{
			errors.New("shapes: " + errData.Message),
		}
		return
	}

	if e.create {
		err = shpe.Insert(db)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"shape_id": shpe.Id.Hex(),
		}).Info("tui: Shape created")
	} else {
		err = shpe.CommitFields(db, commitFields)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"shape_id": shpe.Id.Hex(),
		}).Info("tui: Shape settings saved")
	}

	err = event.PublishDispatch(db, "shape.change")
	if err != nil {
		return
	}

	return
}

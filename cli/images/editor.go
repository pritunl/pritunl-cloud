package images

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/cli/form"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/image"
	"github.com/sirupsen/logrus"
)

// commitFields are the fields written by the admin image handler.
var commitFields = set.NewSet(
	"name",
	"comment",
	"organization",
)

// editor is the image settings form mirroring the inputs of the
// detailed image view in the web interface.
type editor struct {
	img  *image.Image
	form *form.Form

	name    *form.Text
	comment *form.Area
}

func newEditor(img *image.Image) *editor {
	e := &editor{
		img: img,
	}

	e.name = form.NewText("Name", "Enter name", img.Name)
	e.comment = form.NewArea("Comment", "Image comment", img.Comment, 4)

	e.form = form.New(
		e.name,
		e.comment,
	)

	return e
}

func (e *editor) Form() *form.Form {
	return e.form
}

// Save applies the form to the current image the same way as the admin
// image handler, the organization is kept.
func (e *editor) Save(db *database.Database) (err error) {
	img, err := image.Get(db, e.img.Id)
	if err != nil {
		return
	}

	img.Name = e.name.Value()
	img.Comment = e.comment.Value()

	errData, err := img.Validate(db)
	if err != nil {
		return
	}
	if errData != nil {
		err = &errortypes.ParseError{
			errors.New("images: " + errData.Message),
		}
		return
	}

	err = img.CommitFields(db, commitFields)
	if err != nil {
		return
	}

	logrus.WithFields(logrus.Fields{
		"image_id": img.Id.Hex(),
	}).Info("tui: Image settings saved")

	err = event.PublishDispatch(db, "image.change")
	if err != nil {
		return
	}

	return
}

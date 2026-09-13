package storages

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/cli/form"
	"github.com/pritunl/pritunl-cloud/cli/widget"
	"github.com/pritunl/pritunl-cloud/data"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/storage"
	"github.com/sirupsen/logrus"
)

// commitFields are the fields written by the admin storage handler.
var commitFields = set.NewSet(
	"name",
	"comment",
	"type",
	"endpoint",
	"bucket",
	"access_key",
	"secret_key",
	"insecure",
)

var typeOptions = []widget.SelectOption{
	{Label: "Public", Value: storage.Public},
	{Label: "Private", Value: storage.Private},
	{Label: "Web", Value: storage.Web},
}

// editor is the storage settings form mirroring the inputs of the
// detailed storage view in the web interface.
type editor struct {
	store  *storage.Storage
	create bool
	form   *form.Form

	name      *form.Text
	comment   *form.Area
	endpoint  *form.Text
	bucket    *form.Text
	typ       *form.Select
	accessKey *form.Text
	secretKey *form.Text
	ssl       *form.Toggle
}

func newEditor(store *storage.Storage) *editor {
	e := &editor{
		store: store,
	}

	e.name = form.NewText("Name", "Enter name", store.Name)
	e.comment = form.NewArea("Comment", "Storage comment", store.Comment, 4)
	e.endpoint = form.NewText("Endpoint", "Enter endpoint", store.Endpoint)
	e.bucket = form.NewText("Bucket", "Enter bucket", store.Bucket)
	e.typ = form.NewSelect("Type", typeOptions, store.Type)

	// Web storages are read without credentials
	web := func() bool {
		return e.typ.Value() == storage.Web
	}
	e.accessKey = form.NewText("Access Key", "Enter access key",
		store.AccessKey)
	e.accessKey.Hide = web
	e.secretKey = form.NewText("Secret Key", "Enter secret key",
		store.SecretKey)
	e.secretKey.Hide = web

	e.ssl = form.NewToggle("SSL Connection", !store.Insecure)

	e.form = form.New(
		e.name,
		e.comment,
		e.endpoint,
		e.bucket,
		e.typ,
		e.accessKey,
		e.secretKey,
		e.ssl,
	)

	return e
}

func (e *editor) Form() *form.Form {
	return e.form
}

// apply sets the form values on the storage.
func (e *editor) apply(store *storage.Storage) {
	store.Name = e.name.Value()
	store.Comment = e.comment.Value()
	store.Type = e.typ.Value()
	store.Endpoint = e.endpoint.Value()
	store.Bucket = e.bucket.Value()
	store.AccessKey = e.accessKey.Value()
	store.SecretKey = e.secretKey.Value()
	store.Insecure = !e.ssl.Value()
}

// syncBackground syncs the images of the storage in the background
// like the admin storage handlers after every commit.
func syncBackground(store *storage.Storage) {
	go func() {
		db := database.GetDatabase()
		defer db.Close()

		err := data.Sync(db, store)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"storage_id": store.Id.Hex(),
				"error":      err,
			}).Error("tui: Failed to sync storage")
		}

		_ = event.PublishDispatch(db, "image.change")
	}()
}

// Save applies the form to the current storage the same way as the
// admin storage handler, or inserts a new storage, and syncs the images
// in the background.
func (e *editor) Save(db *database.Database) (err error) {
	var store *storage.Storage
	if e.create {
		store = &storage.Storage{}
	} else {
		store, err = storage.Get(db, e.store.Id)
		if err != nil {
			return
		}
	}

	e.apply(store)

	errData, err := store.Validate(db)
	if err != nil {
		return
	}
	if errData != nil {
		err = &errortypes.ParseError{
			errors.New("storages: " + errData.Message),
		}
		return
	}

	if e.create {
		err = store.Insert(db)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"storage_id": store.Id.Hex(),
		}).Info("tui: Storage created")
	} else {
		err = store.CommitFields(db, commitFields)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"storage_id": store.Id.Hex(),
		}).Info("tui: Storage settings saved")
	}

	syncBackground(store)

	err = event.PublishDispatch(db, "storage.change")
	if err != nil {
		return
	}

	return
}

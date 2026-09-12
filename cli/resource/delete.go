package resource

import (
	"fmt"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/relations"
	"github.com/sirupsen/logrus"
)

// DeleteKey is the key of the delete action of every resource.
const DeleteKey = "d"

// DeleteAction returns the red delete button of an expanded resource,
// the confirm dialog stands in for the typed confirmation of the web
// interface. Run does what the admin delete handler does.
func DeleteAction(singular, name string,
	run func(db *database.Database) error) Action {

	return Action{
		Key:     DeleteKey,
		Label:   "Delete",
		Status:  "Deleted",
		Danger:  true,
		Confirm: fmt.Sprintf("Permanently delete the %s %s?", singular, name),
		Run:     run,
	}
}

// DeleteError wraps a handler error message so the dialog shows it.
func DeleteError(singular, message string) error {
	return &errortypes.ParseError{
		errors.New(singular + "s: " + message),
	}
}

// DeleteRelated returns the delete run of resources the admin handlers
// remove after the relations check, the kind is the relations registry
// kind, the remove function is the package Remove and the event the
// change event dispatched afterwards.
func DeleteRelated(kind, singular, eventName string, id bson.ObjectID,
	remove func(db *database.Database, id bson.ObjectID) error) func(
	db *database.Database) error {

	return func(db *database.Database) (err error) {
		errData, err := relations.CanDelete(db, kind, id)
		if err != nil {
			return
		}
		if errData != nil {
			err = DeleteError(singular, errData.Message)
			return
		}

		return Delete(singular, eventName, id, remove)(db)
	}
}

// Delete returns the delete run of resources the admin handlers remove
// without further checks.
func Delete(singular, eventName string, id bson.ObjectID,
	remove func(db *database.Database, id bson.ObjectID) error) func(
	db *database.Database) error {

	return func(db *database.Database) (err error) {
		err = remove(db, id)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			singular + "_id": id.Hex(),
		}).Info("tui: " + singular + " deleted")

		err = event.PublishDispatch(db, eventName)
		if err != nil {
			return
		}

		return
	}
}

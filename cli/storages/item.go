package storages

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/cli/widget"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/storage"
	"github.com/sirupsen/logrus"
)

// Item is a storage card.
type Item struct {
	store *storage.Storage
}

func (i *Item) Storage() *storage.Storage {
	return i.store
}

func (i *Item) Id() string {
	return i.store.Id.Hex()
}

func (i *Item) Name() string {
	return i.store.Name
}

func (i *Item) Tag() string {
	return i.store.Id.Hex()
}

func typeLabel(typ string) string {
	for _, opt := range typeOptions {
		if opt.Value == typ {
			return opt.Label
		}
	}
	return typ
}

func (i *Item) Fields() []resource.Field {
	return []resource.Field{
		{
			Label: "Type",
			Value: typeLabel(i.store.Type),
		},
		{
			Label: "Endpoint",
			Value: i.store.Endpoint,
		},
		{
			Label: "Bucket",
			Value: i.store.Bucket,
		},
	}
}

// Info mirrors the fields of the detailed storage view.
func (i *Item) Info() []widget.InfoField {
	return []widget.InfoField{
		{"ID", i.store.Id.Hex()},
	}
}

// sync starts the image sync of the storage in the background the same
// way as the sync button of the web interface, which commits the
// storage unchanged.
func sync(storeId bson.ObjectID) func(db *database.Database) error {
	return func(db *database.Database) (err error) {
		store, err := storage.Get(db, storeId)
		if err != nil {
			return
		}

		syncBackground(store)

		logrus.WithFields(logrus.Fields{
			"storage_id": store.Id.Hex(),
		}).Info("tui: Storage sync started")

		err = event.PublishDispatch(db, "storage.change")
		if err != nil {
			return
		}

		return
	}
}

// Actions mirrors the sync and delete buttons of the detailed storage
// view.
func (i *Item) Actions() []resource.Action {
	return []resource.Action{
		{
			Key:     "s",
			Label:   "Sync",
			Status:  "Syncing",
			Confirm: "Sync the images of the storage " + i.store.Name + "?",
			Run:     sync(i.store.Id),
		},
		resource.DeleteAction("storage", i.store.Name,
			resource.Delete("storage", "storage.change", i.store.Id,
				storage.Remove)),
	}
}

func (i *Item) Editor() resource.Editor {
	return newEditor(i.store)
}

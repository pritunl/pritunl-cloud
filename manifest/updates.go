package manifest

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/telemetry"
)

const (
	UpdatesType = "updates"

	NodeVariant     = "node"
	InstanceVariant = "instance"
)

type Updates struct {
	Id       bson.ObjectID       `bson:"_id,omitempty" json:"id"`
	Type     string              `bson:"type" json:"type"`
	Resource bson.ObjectID       `bson:"resource" json:"resource"`
	Variant  string              `bson:"variant" json:"variant"`
	Updates  []*telemetry.Update `bson:"updates" json:"updates"`
}

func FindUpdates(db *database.Database) (cursor *UpdatesCursor, err error) {
	cur, err := findEntries(db, &bson.M{
		"type": UpdatesType,
	})
	if err != nil {
		return
	}

	cursor = &UpdatesCursor{
		cursor: cur,
	}

	return
}

type UpdatesCursor struct {
	cursor *Cursor
}

func (c *UpdatesCursor) Next() bool {
	return c.cursor.Next()
}

func (c *UpdatesCursor) Decode() (updt *Updates, err error) {
	updt = &Updates{}

	err = c.cursor.Decode(updt)
	if err != nil {
		return
	}

	return
}

func (c *UpdatesCursor) Err() (err error) {
	err = c.cursor.Err()
	return
}

func (c *UpdatesCursor) Close() {
	c.cursor.Close()
}

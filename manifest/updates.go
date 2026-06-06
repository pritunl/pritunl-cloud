package manifest

import (
	"time"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/mongo-go-driver/v2/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/telemetry"
)

const (
	UpdatesType = "updates"

	NodeVariant     = "node"
	InstanceVariant = "instance"
)

type Updates struct {
	Id           bson.ObjectID       `bson:"_id,omitempty" json:"id"`
	Type         string              `bson:"type" json:"type"`
	Organization bson.ObjectID       `bson:"organization" json:"organization"`
	Resource     bson.ObjectID       `bson:"resource" json:"resource"`
	Timestamp    time.Time           `bson:"timestamp" json:"timestamp"`
	Variant      string              `bson:"variant" json:"variant"`
	Updates      []*telemetry.Update `bson:"updates" json:"updates"`
}

func (u *Updates) Upsert(db *database.Database) (err error) {
	coll := db.Manifests()

	u.Timestamp = time.Now()

	_, err = coll.UpdateOne(db, &bson.M{
		"type":     u.Type,
		"resource": u.Resource,
	}, &bson.M{
		"$set": u,
	}, options.UpdateOne().SetUpsert(true))
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func upsertUpdates(db *database.Database, variant string,
	resource, orgId bson.ObjectID, updates []*telemetry.Update) (err error) {

	entry := &Updates{
		Type:         UpdatesType,
		Resource:     resource,
		Organization: orgId,
		Variant:      variant,
		Updates:      updates,
	}

	err = entry.Upsert(db)
	if err != nil {
		return
	}

	return
}

func UpsertInstanceUpdates(db *database.Database,
	instId, orgId bson.ObjectID, updates []*telemetry.Update) (err error) {

	err = upsertUpdates(db, InstanceVariant, instId, orgId, updates)
	if err != nil {
		return
	}

	return
}

func UpsertNodeUpdates(db *database.Database,
	instId, orgId bson.ObjectID, updates []*telemetry.Update) (err error) {

	err = upsertUpdates(db, NodeVariant, instId, orgId, updates)
	if err != nil {
		return
	}

	return
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

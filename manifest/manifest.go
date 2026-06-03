package manifest

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/mongo-go-driver/v2/mongo"
	"github.com/pritunl/pritunl-cloud/database"
)

type Entry struct {
	Id       bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Type     string        `bson:"type" json:"type"`
	Resource bson.ObjectID `bson:"resource" json:"resource"`
}

type Cursor struct {
	db     *database.Database
	cursor *mongo.Cursor
}

func (c *Cursor) Next() bool {
	return c.cursor.Next(c.db)
}

func (c *Cursor) Decode(data interface{}) (err error) {
	err = c.cursor.Decode(data)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func (c *Cursor) Err() (err error) {
	err = c.cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func (c *Cursor) Close() {
	_ = c.cursor.Close(c.db)
}

func findEntries(db *database.Database, query *bson.M) (
	cursor *Cursor, err error) {

	coll := db.Manifests()

	cur, err := coll.Find(db, query)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	cursor = &Cursor{
		db:     db,
		cursor: cur,
	}

	return
}

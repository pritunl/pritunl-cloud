package block

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"gopkg.in/mgo.v2/bson"
)

type Block struct {
	Id        bson.ObjectId `bson:"_id,omitempty" json:"id"`
	Name      string        `bson:"name" json:"name"`
	Addresses []string      `bson:"addresses" json:"addresses"`
	Excludes  []string      `bson:"excludes" json:"excludes"`
	Netmask   string        `bson:"netmask" json:"netmask"`
	Gateway   string        `bson:"gateway" json:"gateway"`
}

func (b *Block) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	return
}

func (b *Block) Commit(db *database.Database) (err error) {
	coll := db.Blocks()

	err = coll.Commit(b.Id, b)
	if err != nil {
		return
	}

	return
}

func (b *Block) CommitFields(db *database.Database, fields set.Set) (
	err error) {

	coll := db.Blocks()

	err = coll.CommitFields(b.Id, b, fields)
	if err != nil {
		return
	}

	return
}

func (b *Block) Insert(db *database.Database) (err error) {
	coll := db.Blocks()

	if b.Id != "" {
		err = &errortypes.DatabaseError{
			errors.New("block: Block already exists"),
		}
		return
	}

	err = coll.Insert(b)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

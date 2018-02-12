package organization

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"gopkg.in/mgo.v2/bson"
)

type Organization struct {
	Id    bson.ObjectId `bson:"_id,omitempty" json:"id"`
	Roles []string      `bson:"roles" json:"roles"`
	Name  string        `bson:"name" json:"name"`
}

func (d *Organization) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	if d.Roles == nil {
		d.Roles = []string{}
	}

	return
}

func (d *Organization) Commit(db *database.Database) (err error) {
	coll := db.Organizations()

	err = coll.Commit(d.Id, d)
	if err != nil {
		return
	}

	return
}

func (d *Organization) CommitFields(db *database.Database, fields set.Set) (
	err error) {

	coll := db.Organizations()

	err = coll.CommitFields(d.Id, d, fields)
	if err != nil {
		return
	}

	return
}

func (c *Organization) Insert(db *database.Database) (err error) {
	coll := db.Organizations()

	if c.Id != "" {
		err = &errortypes.DatabaseError{
			errors.New("organization: Organization already exists"),
		}
		return
	}

	err = coll.Insert(c)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

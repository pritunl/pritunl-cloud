package organization

import (
	"context"
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

type Organization struct {
	Id    primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Roles []string           `bson:"roles" json:"roles"`
	Name  string             `bson:"name" json:"name"`
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

	if !c.Id.IsZero() {
		err = &errortypes.DatabaseError{
			errors.New("organization: Organization already exists"),
		}
		return
	}

	_, err = coll.InsertOne(context.Background(), c)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

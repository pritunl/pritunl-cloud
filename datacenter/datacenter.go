package datacenter

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"gopkg.in/mgo.v2/bson"
)

type Datacenter struct {
	Id            bson.ObjectId   `bson:"_id,omitempty" json:"id"`
	Organizations []bson.ObjectId `bson:"organizations" json:"organizations"`
	Name          string          `bson:"name" json:"name"`
}

func (d *Datacenter) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	if d.Organizations == nil {
		d.Organizations = []bson.ObjectId{}
	}

	return
}

func (d *Datacenter) Commit(db *database.Database) (err error) {
	coll := db.Datacenters()

	err = coll.Commit(d.Id, d)
	if err != nil {
		return
	}

	return
}

func (d *Datacenter) CommitFields(db *database.Database, fields set.Set) (
	err error) {

	coll := db.Datacenters()

	err = coll.CommitFields(d.Id, d, fields)
	if err != nil {
		return
	}

	return
}

func (d *Datacenter) Insert(db *database.Database) (err error) {
	coll := db.Datacenters()

	if d.Id != "" {
		err = &errortypes.DatabaseError{
			errors.New("datacenter: Datacenter already exists"),
		}
		return
	}

	err = coll.Insert(d)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

package vpc

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"gopkg.in/mgo.v2/bson"
)

type Vpc struct {
	Id           bson.ObjectId `bson:"_id,omitempty" json:"id"`
	Name         string        `bson:"name" json:"name"`
	VpcId        int           `bson:"vpc_id" json:"vpc_id"`
	Network      string        `bson:"network" json:"network"`
	Organization bson.ObjectId `bson:"organization" json:"organization"`
	Datacenter   bson.ObjectId `bson:"datacenter" json:"datacenter"`
}

func (v *Vpc) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	return
}

func (v *Vpc) Commit(db *database.Database) (err error) {
	coll := db.Vpcs()

	err = coll.Commit(v.Id, v)
	if err != nil {
		return
	}

	return
}

func (v *Vpc) CommitFields(db *database.Database, fields set.Set) (
	err error) {

	coll := db.Vpcs()

	err = coll.CommitFields(v.Id, v, fields)
	if err != nil {
		return
	}

	return
}

func (v *Vpc) Insert(db *database.Database) (err error) {
	coll := db.Vpcs()

	if v.Id != "" {
		err = &errortypes.DatabaseError{
			errors.New("firewall: Vpc already exists"),
		}
		return
	}

	err = coll.Insert(v)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

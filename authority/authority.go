package authority

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"gopkg.in/mgo.v2/bson"
)

type Authority struct {
	Id             bson.ObjectId `bson:"_id,omitempty" json:"id"`
	Name           string        `bson:"name" json:"name"`
	Organization   bson.ObjectId `bson:"organization,omitempty" json:"organization"`
	AuthorityRoles []string      `bson:"authority_roles" json:"authority_roles"`
}

func (f *Authority) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	return
}

func (f *Authority) Commit(db *database.Database) (err error) {
	coll := db.Authorities()

	err = coll.Commit(f.Id, f)
	if err != nil {
		return
	}

	return
}

func (f *Authority) CommitFields(db *database.Database, fields set.Set) (
	err error) {

	coll := db.Authorities()

	err = coll.CommitFields(f.Id, f, fields)
	if err != nil {
		return
	}

	return
}

func (f *Authority) Insert(db *database.Database) (err error) {
	coll := db.Authorities()

	if f.Id != "" {
		err = &errortypes.DatabaseError{
			errors.New("firewall: Authority already exists"),
		}
		return
	}

	err = coll.Insert(f)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

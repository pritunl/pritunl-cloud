package domain

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"gopkg.in/mgo.v2/bson"
)

type Domain struct {
	Id           bson.ObjectId `bson:"_id,omitempty" json:"id"`
	Name         string        `bson:"name" json:"name"`
	Organization bson.ObjectId `bson:"organization,omitempty" json:"organization"`
	Type         string        `bson:"type" json:"type"`
	AwsId        string        `bson:"aws_id" json:"aws_id"`
	AwsSecret    string        `bson:"aws_secret" json:"aws_secret"`
}

func (d *Domain) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	if d.Organization == "" {
		errData = &errortypes.ErrorData{
			Error:   "organization_required",
			Message: "Missing required organization",
		}
		return
	}

	if d.Type != Route53 {
		d.AwsId = ""
		d.AwsSecret = ""
	}

	return
}

func (d *Domain) Commit(db *database.Database) (err error) {
	coll := db.Domains()

	err = coll.Commit(d.Id, d)
	if err != nil {
		return
	}

	return
}

func (d *Domain) CommitFields(db *database.Database, fields set.Set) (
	err error) {

	coll := db.Domains()

	err = coll.CommitFields(d.Id, d, fields)
	if err != nil {
		return
	}

	return
}

func (d *Domain) Insert(db *database.Database) (err error) {
	coll := db.Domains()

	if d.Id != "" {
		err = &errortypes.DatabaseError{
			errors.New("domain: Domain already exists"),
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

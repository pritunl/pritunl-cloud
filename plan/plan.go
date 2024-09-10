package plan

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

type Plan struct {
	Id           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name         string             `bson:"name" json:"name"`
	Comment      string             `bson:"comment" json:"comment"`
	Organization primitive.ObjectID `bson:"organization,omitempty" json:"organization"`
	Type         string             `bson:"type" json:"type"`
}

func (p *Plan) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	if p.Organization.IsZero() {
		errData = &errortypes.ErrorData{
			Error:   "organization_required",
			Message: "Missing required organization",
		}
		return
	}

	switch p.Type {
	case Rolling, "":
		p.Type = Rolling
		break
	case Recreate:
		break
	default:
		errData = &errortypes.ErrorData{
			Error:   "type_invalid",
			Message: "Type invalid",
		}
		return
	}

	return
}

func (p *Plan) Commit(db *database.Database) (err error) {
	coll := db.Plans()

	err = coll.Commit(p.Id, p)
	if err != nil {
		return
	}

	return
}

func (p *Plan) CommitFields(db *database.Database, fields set.Set) (
	err error) {

	coll := db.Plans()

	err = coll.CommitFields(p.Id, p, fields)
	if err != nil {
		return
	}

	return
}

func (p *Plan) Insert(db *database.Database) (err error) {
	coll := db.Plans()

	if !p.Id.IsZero() {
		err = &errortypes.DatabaseError{
			errors.New("domain: Plan already exists"),
		}
		return
	}

	_, err = coll.InsertOne(db, p)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

package authority

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/utils"
)

type Authority struct {
	Id           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name         string             `bson:"name" json:"name"`
	Comment      string             `bson:"comment" json:"comment"`
	Type         string             `bson:"type" json:"type"`
	Organization primitive.ObjectID `bson:"organization" json:"organization"`
	NetworkRoles []string           `bson:"network_roles" json:"network_roles"`
	Key          string             `bson:"key" json:"key"`
	Roles        []string           `bson:"roles" json:"roles"`
	Certificate  string             `bson:"certificate" json:"certificate"`
}

func (f *Authority) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	f.Name = utils.FilterName(f.Name)

	if f.Roles == nil {
		f.Roles = []string{}
	}

	if f.Type == "" {
		f.Type = SshKey
	}

	switch f.Type {
	case SshKey:
		f.Roles = []string{}
		f.Certificate = ""
		break
	case SshCertificate:
		f.Key = ""
		break
	}

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

	if !f.Id.IsZero() {
		err = &errortypes.DatabaseError{
			errors.New("authority: Authority already exists"),
		}
		return
	}

	resp, err := coll.InsertOne(db, f)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	f.Id = resp.InsertedID.(primitive.ObjectID)

	return
}

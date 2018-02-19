package storage

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"gopkg.in/mgo.v2/bson"
)

type Storage struct {
	Id        bson.ObjectId `bson:"_id,omitempty" json:"id"`
	Name      string        `bson:"name" json:"name"`
	Type      string        `bson:"type" json:"type"`
	Endpoint  string        `bson:"endpoint" json:"endpoint"`
	Bucket    string        `bson:"bucket" json:"bucket"`
	AccessKey string        `bson:"access_key" json:"access_key"`
	SecretKey string        `bson:"secret_key" json:"secret_key"`
	Insecure  bool          `bson:"insecure" json:"insecure"`
}

func (s *Storage) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	if s.Type == "" {
		s.Type = Public
	}

	return
}

func (s *Storage) Commit(db *database.Database) (err error) {
	coll := db.Storages()

	err = coll.Commit(s.Id, s)
	if err != nil {
		return
	}

	return
}

func (s *Storage) CommitFields(db *database.Database, fields set.Set) (
	err error) {

	coll := db.Storages()

	err = coll.CommitFields(s.Id, s, fields)
	if err != nil {
		return
	}

	return
}

func (s *Storage) Insert(db *database.Database) (err error) {
	coll := db.Storages()

	if s.Id != "" {
		err = &errortypes.DatabaseError{
			errors.New("storage: Storage already exists"),
		}
		return
	}

	err = coll.Insert(s)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

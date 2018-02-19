package image

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"gopkg.in/mgo.v2/bson"
)

type Image struct {
	Id           bson.ObjectId `bson:"_id,omitempty" json:"id"`
	Name         string        `bson:"name" json:"name"`
	Organization bson.ObjectId `bson:"organization" json:"organization"`
	Type         string        `bson:"type" json:"type"`
	Storage      bson.ObjectId `bson:"storage" json:"storage"`
	Key          string        `bson:"key" json:"key"`
	Etag         string        `bson:"etag" json:"etag"`
}

func (i *Image) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	return
}

func (i *Image) Json() {
	if i.Name == "" {
		i.Name = i.Key
	}
}

func (i *Image) Commit(db *database.Database) (err error) {
	coll := db.Images()

	err = coll.Commit(i.Id, i)
	if err != nil {
		return
	}

	return
}

func (i *Image) CommitFields(db *database.Database, fields set.Set) (
	err error) {

	coll := db.Images()

	err = coll.CommitFields(i.Id, i, fields)
	if err != nil {
		return
	}

	return
}

func (i *Image) Insert(db *database.Database) (err error) {
	coll := db.Images()

	if i.Id != "" {
		err = &errortypes.DatabaseError{
			errors.New("image: Image already exists"),
		}
		return
	}

	err = coll.Insert(i)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func (i *Image) Upsert(db *database.Database) (err error) {
	coll := db.Images()

	_, err = coll.Upsert(&bson.M{
		"storage": i.Storage,
		"key":     i.Key,
	}, &bson.M{
		"$set": &bson.M{
			"storage": i.Storage,
			"key":     i.Key,
			"etag":    i.Etag,
			"type":    i.Type,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

package image

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"gopkg.in/mgo.v2/bson"
	"strings"
	"time"
)

type Image struct {
	Id           bson.ObjectId `bson:"_id,omitempty" json:"id"`
	Disk         bson.ObjectId `bson:"disk,omitempty" json:"disk"`
	Name         string        `bson:"name" json:"name"`
	Organization bson.ObjectId `bson:"organization" json:"organization"`
	Signed       bool          `bson:"signed" json:"signed"`
	Type         string        `bson:"type" json:"type"`
	Storage      bson.ObjectId `bson:"storage" json:"storage"`
	Key          string        `bson:"key" json:"key"`
	LastModified time.Time     `bson:"last_modified" json:"last_modified"`
	StorageClass string        `bson:"storage_class" json:"storage_class"`
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
			"disk":          i.Disk,
			"name":          i.Name,
			"organization":  i.Organization,
			"signed":        i.Signed,
			"type":          i.Type,
			"storage":       i.Storage,
			"key":           i.Key,
			"last_modified": i.LastModified,
			"storage_class": i.StorageClass,
			"etag":          i.Etag,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func (i *Image) Sync(db *database.Database) (err error) {
	coll := db.Images()

	if strings.HasPrefix(i.Key, "backup/") ||
		strings.HasPrefix(i.Key, "snapshot/") {

		err = coll.Update(&bson.M{
			"storage": i.Storage,
			"key":     i.Key,
		}, &bson.M{
			"$set": &bson.M{
				"storage":       i.Storage,
				"key":           i.Key,
				"signed":        i.Signed,
				"type":          i.Type,
				"etag":          i.Etag,
				"last_modified": i.LastModified,
				"storage_class": i.StorageClass,
			},
		})
		if err != nil {
			err = database.ParseError(err)
			if _, ok := err.(*database.NotFoundError); ok {
				err = &LostImageError{
					errors.Wrap(err, "image: Lost image"),
				}
			}
			return
		}
	} else {
		_, err = coll.Upsert(&bson.M{
			"storage": i.Storage,
			"key":     i.Key,
		}, &bson.M{
			"$set": &bson.M{
				"storage":       i.Storage,
				"key":           i.Key,
				"signed":        i.Signed,
				"type":          i.Type,
				"etag":          i.Etag,
				"last_modified": i.LastModified,
			},
		})
		if err != nil {
			err = database.ParseError(err)
			return
		}
	}

	return
}

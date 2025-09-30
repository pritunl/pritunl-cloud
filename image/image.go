package image

import (
	"strings"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/mongo-go-driver/v2/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/deployment"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/utils"
)

type Image struct {
	Id           bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Disk         bson.ObjectID `bson:"disk" json:"disk"`
	Name         string        `bson:"name" json:"name"`
	Release      string        `bson:"release" json:"release"`
	Build        string        `bson:"build" json:"build"`
	Comment      string        `bson:"comment" json:"comment"`
	Deployment   bson.ObjectID `bson:"deployment" json:"deployment"`
	Organization bson.ObjectID `bson:"organization" json:"organization"`
	Signed       bool          `bson:"signed" json:"signed"`
	Type         string        `bson:"type" json:"type"`
	SystemType   string        `bson:"system_type" json:"system_type"`
	Firmware     string        `bson:"firmware" json:"firmware"`
	Storage      bson.ObjectID `bson:"storage" json:"storage"`
	Key          string        `bson:"key" json:"key"`
	LastModified time.Time     `bson:"last_modified" json:"last_modified"`
	StorageClass string        `bson:"storage_class" json:"storage_class"`
	Hash         string        `bson:"hash" json:"hash"`
	Etag         string        `bson:"etag" json:"etag"`
	Tags         []string      `bson:"-" json:"tags"`
}

type Completion struct {
	Id           bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name         string        `bson:"name" json:"name"`
	Release      string        `bson:"release" json:"release"`
	Build        string        `bson:"build" json:"build"`
	Organization bson.ObjectID `bson:"organization" json:"organization"`
	Deployment   bson.ObjectID `bson:"deployment" json:"deployment"`
	Type         string        `bson:"type" json:"type"`
	Firmware     string        `bson:"firmware" json:"firmware"`
	Key          string        `bson:"key" json:"key"`
	Storage      bson.ObjectID `bson:"storage" json:"storage"`
	Tags         []string      `bson:"-" json:"tags"`
}

func (i *Image) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	i.Name = utils.FilterName(i.Name)

	if i.Firmware == "" {
		i.Firmware = Uefi
	}

	switch i.SystemType {
	case Linux:
		break
	case LinuxLegacy:
		break
	case LinuxUnsigned:
		break
	case Bsd:
		break
	default:
		errData = &errortypes.ErrorData{
			Error:   "invalid_system_type",
			Message: "Image system type invalid",
		}
		return
	}

	return
}

func (i *Image) Parse() {
	if i.Name == "" {
		i.Name = i.Key
	}

	if i.Signed {
		i.Name, i.Release, i.Build = ParseImageName(i.Key)
	}
}

func (i *Image) GetSystemType() string {
	if i.SystemType != "" {
		return i.SystemType
	}

	name := strings.ToLower(i.Name)
	if strings.Contains(name, "bsd") {
		return Bsd
	}

	if strings.Contains(name, "alpinelinux") {
		return LinuxUnsigned
	}

	return Linux
}

func (i *Image) Json() {
	i.Parse()
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

	_, err = coll.InsertOne(db, i)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func (i *Image) Upsert(db *database.Database) (err error) {
	coll := db.Images()

	fields := bson.M{
		"name":          i.Name,
		"deployment":    i.Deployment,
		"organization":  i.Organization,
		"disk":          i.Disk,
		"signed":        i.Signed,
		"type":          i.Type,
		"system_type":   i.SystemType,
		"firmware":      i.Firmware,
		"storage":       i.Storage,
		"key":           i.Key,
		"last_modified": i.LastModified,
		"storage_class": i.StorageClass,
		"hash":          i.Hash,
		"etag":          i.Etag,
	}

	resp, err := coll.UpdateOne(
		db,
		&bson.M{
			"storage": i.Storage,
			"key":     i.Key,
		},
		&bson.M{
			"$set": fields,
		},
		options.UpdateOne().SetUpsert(true),
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	if resp.UpsertedID != nil {
		i.Id = resp.UpsertedID.(bson.ObjectID)
	}

	return
}

func (i *Image) Sync(db *database.Database) (err error) {
	coll := db.Images()

	i.Parse()

	if strings.HasPrefix(i.Key, "backup/") ||
		strings.HasPrefix(i.Key, "snapshot/") {

		resp, e := coll.UpdateOne(
			db,
			&bson.M{
				"storage": i.Storage,
				"key":     i.Key,
			},
			&bson.M{
				"$set": &bson.M{
					"organization":  bson.NilObjectID,
					"release":       i.Release,
					"build":         i.Build,
					"storage":       i.Storage,
					"key":           i.Key,
					"signed":        i.Signed,
					"type":          i.Type,
					"firmware":      i.Firmware,
					"etag":          i.Etag,
					"last_modified": i.LastModified,
					"storage_class": i.StorageClass,
				},
				"$setOnInsert": &bson.M{
					"name":       i.Name,
					"disk":       bson.NilObjectID,
					"deployment": bson.NilObjectID,
				},
			},
		)
		if e != nil {
			err = database.ParseError(e)
			if _, ok := err.(*database.NotFoundError); ok {
				err = &LostImageError{
					errors.Wrap(err, "image: Lost image"),
				}
			}
			return
		}

		if resp.UpsertedID != nil {
			i.Id = resp.UpsertedID.(bson.ObjectID)
		}
	} else {
		resp, e := coll.UpdateOne(
			db,
			&bson.M{
				"storage": i.Storage,
				"key":     i.Key,
			},
			&bson.M{
				"$set": &bson.M{
					"organization":  bson.NilObjectID,
					"name":          i.Name,
					"release":       i.Release,
					"build":         i.Build,
					"storage":       i.Storage,
					"key":           i.Key,
					"signed":        i.Signed,
					"type":          i.Type,
					"firmware":      i.Firmware,
					"etag":          i.Etag,
					"last_modified": i.LastModified,
				},
				"$setOnInsert": &bson.M{
					"disk":       bson.NilObjectID,
					"deployment": bson.NilObjectID,
				},
			},
			options.UpdateOne().SetUpsert(true),
		)
		if e != nil {
			err = database.ParseError(e)
			return
		}

		if resp.UpsertedID != nil {
			i.Id = resp.UpsertedID.(bson.ObjectID)
		}
	}

	return
}

func (i *Image) Remove(db *database.Database) (err error) {
	if !i.Deployment.IsZero() {
		err = deployment.Remove(db, i.Deployment)
		if err != nil {
			return
		}
	}

	err = Remove(db, i.Id)
	if err != nil {
		return
	}

	return
}

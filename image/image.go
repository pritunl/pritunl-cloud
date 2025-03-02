package image

import (
	"fmt"
	"strings"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/mongo-go-driver/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/deployment"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/utils"
)

type Image struct {
	Id           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Disk         primitive.ObjectID `bson:"disk,omitempty" json:"disk"`
	Name         string             `bson:"name" json:"name"`
	Comment      string             `bson:"comment" json:"comment"`
	Deployment   primitive.ObjectID `bson:"deployment,omitempty" json:"deployment"`
	Organization primitive.ObjectID `bson:"organization,omitempty" json:"organization"`
	Signed       bool               `bson:"signed" json:"signed"`
	Type         string             `bson:"type" json:"type"`
	SystemType   string             `bson:"system_type" json:"system_type"`
	Firmware     string             `bson:"firmware" json:"firmware"`
	Storage      primitive.ObjectID `bson:"storage" json:"storage"`
	Key          string             `bson:"key" json:"key"`
	LastModified time.Time          `bson:"last_modified" json:"last_modified"`
	StorageClass string             `bson:"storage_class" json:"storage_class"`
	Etag         string             `bson:"etag" json:"etag"`
}

func (i *Image) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	i.Name = utils.FilterName(i.Name)

	if i.Firmware == "" {
		i.Firmware = Unknown
	}

	switch i.SystemType {
	case Linux:
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

	if i.Signed && len(i.Name) > 6 {
		names := strings.Split(i.Name[:len(i.Name)-6], "_")
		n := len(names)
		if n == 2 && len(names[1]) >= 4 {
			switch names[0] {
			case "almalinux8":
				i.Name = fmt.Sprintf(
					"almalinux-8-%s-20%s",
					names[1][2:],
					names[1][:2],
				)
				break
			case "almalinux9":
				i.Name = fmt.Sprintf(
					"almalinux-9-%s-20%s",
					names[1][2:],
					names[1][:2],
				)
				break
			case "centos7":
				i.Name = fmt.Sprintf(
					"centos-7-%--20%s",
					names[1][2:],
					names[1][:2],
				)
				break
			case "centos8":
				i.Name = fmt.Sprintf(
					"centos-8-%--20%s",
					names[1][2:],
					names[1][:2],
				)
				break
			case "oraclelinux7":
				i.Name = fmt.Sprintf(
					"oracle-linux-7-%s-20%s",
					names[1][2:],
					names[1][:2],
				)
				break
			case "oraclelinux8":
				i.Name = fmt.Sprintf(
					"oracle-linux-8-%s-20%s",
					names[1][2:],
					names[1][:2],
				)
				break
			case "oraclelinux9":
				i.Name = fmt.Sprintf(
					"oracle-linux-9-%s-20%s",
					names[1][2:],
					names[1][:2],
				)
				break
			case "ubuntu2004":
				i.Name = fmt.Sprintf(
					"ubuntu-2004-%s-20%s",
					names[1][2:],
					names[1][:2],
				)
				break
			case "ubuntu2204":
				i.Name = fmt.Sprintf(
					"ubuntu-2204-%s-20%s",
					names[2][2:],
					names[2][:2],
				)
				break
			case "ubuntu2404":
				i.Name = fmt.Sprintf(
					"ubuntu-2404-%s-20%s",
					names[2][2:],
					names[2][:2],
				)
				break
			case "ubuntu2604":
				i.Name = fmt.Sprintf(
					"ubuntu-2604-%s-20%s",
					names[2][2:],
					names[2][:2],
				)
				break
			}
		} else if n == 3 && names[1] == "efi" && len(names[2]) == 4 {
			switch names[0] {
			case "almalinux8":
				i.Name = fmt.Sprintf(
					"almalinux-8-efi-%s-20%s",
					names[2][2:],
					names[2][:2],
				)
				break
			case "almalinux9":
				i.Name = fmt.Sprintf(
					"almalinux-9-efi-%s-20%s",
					names[2][2:],
					names[2][:2],
				)
				break
			case "centos7":
				i.Name = fmt.Sprintf(
					"centos-7-efi-%s-20%s",
					names[2][2:],
					names[2][:2],
				)
				break
			case "centos8":
				i.Name = fmt.Sprintf(
					"centos-8-efi-%s-20%s",
					names[2][2:],
					names[2][:2],
				)
				break
			case "oraclelinux7":
				i.Name = fmt.Sprintf(
					"oracle-linux-7-efi-%s-20%s",
					names[2][2:],
					names[2][:2],
				)
				break
			case "oraclelinux8":
				i.Name = fmt.Sprintf(
					"oracle-linux-8-efi-%s-20%s",
					names[2][2:],
					names[2][:2],
				)
				break
			case "oraclelinux9":
				i.Name = fmt.Sprintf(
					"oracle-linux-9-efi-%s-20%s",
					names[2][2:],
					names[2][:2],
				)
				break
			case "fedora":
				i.Name = fmt.Sprintf(
					"fedora-efi-%s-20%s",
					names[2][2:],
					names[2][:2],
				)
				break
			case "ubuntu2004":
				i.Name = fmt.Sprintf(
					"ubuntu-2004-efi-%s-20%s",
					names[2][2:],
					names[2][:2],
				)
				break
			case "ubuntu2204":
				i.Name = fmt.Sprintf(
					"ubuntu-2204-efi-%s-20%s",
					names[2][2:],
					names[2][:2],
				)
				break
			case "ubuntu2404":
				i.Name = fmt.Sprintf(
					"ubuntu-2404-efi-%s-20%s",
					names[2][2:],
					names[2][:2],
				)
				break
			case "ubuntu2604":
				i.Name = fmt.Sprintf(
					"ubuntu-2604-efi-%s-20%s",
					names[2][2:],
					names[2][:2],
				)
				break
			case "freebsd":
				i.Name = fmt.Sprintf(
					"freebsd-efi-%s-20%s",
					names[2][2:],
					names[2][:2],
				)
				break
			}
		}
	}
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

	opts := &options.UpdateOptions{}
	opts.SetUpsert(true)
	resp, err := coll.UpdateOne(
		db,
		&bson.M{
			"storage": i.Storage,
			"key":     i.Key,
		},
		&bson.M{
			"$set": &bson.M{
				"disk":          i.Disk,
				"name":          i.Name,
				"deployment":    i.Deployment,
				"organization":  i.Organization,
				"signed":        i.Signed,
				"type":          i.Type,
				"firmware":      i.Firmware,
				"storage":       i.Storage,
				"key":           i.Key,
				"last_modified": i.LastModified,
				"storage_class": i.StorageClass,
				"etag":          i.Etag,
			},
		},
		opts,
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	if resp.UpsertedID != nil {
		i.Id = resp.UpsertedID.(primitive.ObjectID)
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
					"name": i.Name,
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
			i.Id = resp.UpsertedID.(primitive.ObjectID)
		}
	} else {
		opts := &options.UpdateOptions{}
		opts.SetUpsert(true)
		resp, e := coll.UpdateOne(
			db,
			&bson.M{
				"storage": i.Storage,
				"key":     i.Key,
			},
			&bson.M{
				"$set": &bson.M{
					"name":          i.Name,
					"storage":       i.Storage,
					"key":           i.Key,
					"signed":        i.Signed,
					"type":          i.Type,
					"firmware":      i.Firmware,
					"etag":          i.Etag,
					"last_modified": i.LastModified,
				},
			},
			opts,
		)
		if e != nil {
			err = database.ParseError(e)
			return
		}

		if resp.UpsertedID != nil {
			i.Id = resp.UpsertedID.(primitive.ObjectID)
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

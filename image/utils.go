package image

import (
	"context"
	"crypto/md5"
	"fmt"
	"github.com/dropbox/godropbox/container/set"
	"github.com/minio/minio-go"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/mongo-go-driver/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/utils"
	"regexp"
	"time"
)

var (
	etagReg = regexp.MustCompile("[^a-zA-Z0-9]+")
)

func GetEtag(info minio.ObjectInfo) string {
	etag := info.ETag
	if etag == "" {
		modifiedHash := md5.New()
		modifiedHash.Write(
			[]byte(info.LastModified.Format(time.RFC3339)))
		etag = fmt.Sprintf("%x", modifiedHash.Sum(nil))
	}
	return etagReg.ReplaceAllString(etag, "")
}

func Get(db *database.Database, imgId primitive.ObjectID) (
	img *Image, err error) {

	coll := db.Images()
	img = &Image{}

	err = coll.FindOneId(imgId, img)
	if err != nil {
		return
	}

	return
}

func GetOrg(db *database.Database, orgId, imgId primitive.ObjectID) (
	img *Image, err error) {

	coll := db.Images()
	img = &Image{}

	err = coll.FindOne(context.Background(), &bson.M{
		"_id":          imgId,
		"organization": orgId,
	}).Decode(img)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetOrgPublic(db *database.Database, orgId, imgId primitive.ObjectID) (
	img *Image, err error) {

	coll := db.Images()
	img = &Image{}

	err = coll.FindOne(context.Background(), &bson.M{
		"_id": imgId,
		"$or": []*bson.M{
			&bson.M{
				"organization": orgId,
			},
			&bson.M{
				"organization": &bson.M{
					"$exists": false,
				},
			},
		},
	}).Decode(img)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func Distinct(db *database.Database, storeId primitive.ObjectID) (
	keys []string, err error) {

	coll := db.Images()
	keys = []string{}

	keysInf, err := coll.Distinct(context.Background(), "key", &bson.M{
		"storage": storeId,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	for _, keyInf := range keysInf {
		if key, ok := keyInf.(string); ok {
			keys = append(keys, key)
		}
	}

	return
}

func ExistsOrg(db *database.Database, orgId, imgId primitive.ObjectID) (
	exists bool, err error) {

	coll := db.Images()

	n, err := coll.Count(context.Background(), &bson.M{
		"_id": imgId,
		"$or": []*bson.M{
			&bson.M{
				"organization": orgId,
			},
			&bson.M{
				"organization": &bson.M{
					"$exists": false,
				},
			},
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	if n > 0 {
		exists = true
	}

	return
}

func GetAll(db *database.Database, query *bson.M, page, pageCount int64) (
	imgs []*Image, count int64, err error) {

	coll := db.Images()
	imgs = []*Image{}

	count, err = coll.Count(context.Background(), query)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	page = utils.Min64(page, count/pageCount)
	skip := utils.Min64(page*pageCount, count)

	cursor, err := coll.Find(
		context.Background(),
		query,
		&options.FindOptions{
			Sort: &bson.D{
				{"name", 1},
			},
			Skip:  &skip,
			Limit: &pageCount,
		},
	)
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		img := &Image{}
		err = cursor.Decode(img)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		imgs = append(imgs, img)
		img = &Image{}
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllNames(db *database.Database, query *bson.M) (
	images []*Image, err error) {

	coll := db.Images()
	images = []*Image{}

	cursor, err := coll.Find(
		context.Background(),
		query,
		&options.FindOptions{
			Sort: &bson.D{
				{"name", 1},
			},
			Projection: &bson.D{
				{"name", 1},
				{"key", 1},
			},
		},
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		img := &Image{}
		err = cursor.Decode(img)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		images = append(images, img)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllKeys(db *database.Database) (keys set.Set, err error) {
	coll := db.Images()
	keys = set.NewSet()

	cursor, err := coll.Find(
		context.Background(),
		&bson.M{},
		&options.FindOptions{
			Sort: &bson.D{
				{"name", 1},
			},
			Projection: &bson.D{
				{"_id", 1},
				{"etag", 1},
			},
		},
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		img := &Image{}
		err = cursor.Decode(img)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		keys.Add(fmt.Sprintf("%s-%s", img.Id.Hex(), img.Etag))
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func Remove(db *database.Database, imgId primitive.ObjectID) (err error) {
	coll := db.Images()

	_, err = coll.DeleteOne(context.Background(), &bson.M{
		"_id": imgId,
	})
	if err != nil {
		err = database.ParseError(err)
		switch err.(type) {
		case *database.NotFoundError:
			err = nil
		default:
			return
		}
	}

	return
}

func RemoveKeys(db *database.Database, storeId primitive.ObjectID,
	keys []string) (err error) {
	coll := db.Images()

	_, err = coll.DeleteMany(context.Background(), &bson.M{
		"storage": storeId,
		"key": &bson.M{
			"$in": keys,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

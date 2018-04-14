package image

import (
	"crypto/md5"
	"fmt"
	"github.com/dropbox/godropbox/container/set"
	"github.com/minio/minio-go"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/utils"
	"gopkg.in/mgo.v2/bson"
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

func Get(db *database.Database, imgId bson.ObjectId) (
	img *Image, err error) {

	coll := db.Images()
	img = &Image{}

	err = coll.FindOneId(imgId, img)
	if err != nil {
		return
	}

	return
}

func Distinct(db *database.Database, storeId bson.ObjectId) (
	keys []string, err error) {

	coll := db.Images()

	keys = []string{}
	err = coll.Find(&bson.M{
		"storage": storeId,
	}).Distinct("key", &keys)
	if err != nil {
		return
	}

	return
}

func GetAll(db *database.Database, query *bson.M, page, pageCount int) (
	imgs []*Image, count int, err error) {

	coll := db.Images()
	imgs = []*Image{}

	qury := coll.Find(query)

	count, err = qury.Count()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	skip := utils.Min(page*pageCount, utils.Max(0, count-pageCount))

	cursor := qury.Sort("key").Skip(skip).Limit(pageCount).Iter()

	img := &Image{}
	for cursor.Next(img) {
		imgs = append(imgs, img)
		img = &Image{}
	}

	err = cursor.Close()
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

	cursor := coll.Find(query).Sort("key").Select(&bson.M{
		"name": 1,
		"key":  1,
	}).Iter()

	img := &Image{}
	for cursor.Next(img) {
		images = append(images, img)
		img = &Image{}
	}

	err = cursor.Close()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllKeys(db *database.Database) (keys set.Set, err error) {
	coll := db.Images()
	keys = set.NewSet()

	cursor := coll.Find(&bson.M{}).Select(&bson.M{
		"_id":  1,
		"etag": 1,
	}).Iter()

	img := &Image{}
	for cursor.Next(img) {
		keys.Add(fmt.Sprintf("%s-%s", img.Id.Hex(), img.Etag))
	}

	err = cursor.Close()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func Remove(db *database.Database, imgId bson.ObjectId) (err error) {
	coll := db.Images()

	err = coll.Remove(&bson.M{
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

func RemoveKeys(db *database.Database, storeId bson.ObjectId,
	keys []string) (err error) {
	coll := db.Images()

	_, err = coll.RemoveAll(&bson.M{
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

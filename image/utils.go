package image

import (
	"crypto/md5"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/dropbox/godropbox/container/set"
	minio "github.com/minio/minio-go/v7"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/mongo-go-driver/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/utils"
)

var (
	etagReg  = regexp.MustCompile("[^a-zA-Z0-9]+")
	distroRe = regexp.MustCompile(`^([a-z]+)([0-9]*)`)
	dateRe   = regexp.MustCompile(`_(\d{2})(\d{2})(\d{2})?$`)
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

func ParseImageName(key string) (name, release, build string) {
	baseName := strings.TrimSuffix(key, filepath.Ext(key))

	dateMatch := dateRe.FindStringSubmatch(baseName)
	if len(dateMatch) != 3 && len(dateMatch) != 4 {
		name = key
		return
	}
	yearStr, monthStr := dateMatch[1], dateMatch[2]
	build = yearStr + monthStr
	if len(dateMatch) == 4 {
		build += dateMatch[3]
	}

	base := strings.TrimSuffix(baseName, dateMatch[0])
	tokens := strings.Split(base, "_")
	if len(tokens) == 0 {
		name = key
		return
	}

	distroMatch := distroRe.FindStringSubmatch(tokens[0])
	if len(distroMatch) < 2 {
		name = key
		return
	}
	distro := distroMatch[1]
	version := ""
	if len(distroMatch) >= 3 {
		version = distroMatch[2]
	}

	if version == "" {
		name = fmt.Sprintf("%s-%s%s", distro, yearStr, monthStr)
	} else {
		name = fmt.Sprintf("%s%s-%s%s", distro, version, yearStr, monthStr)
	}

	if Releases.Contains(distro + version) {
		release = distro + version
	}

	return
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

func GetKey(db *database.Database, storeId primitive.ObjectID, key string) (
	img *Image, err error) {

	coll := db.Images()
	img = &Image{}

	err = coll.FindOne(db, &bson.M{
		"storage": storeId,
		"key":     key,
	}).Decode(img)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetOrg(db *database.Database, orgId, imgId primitive.ObjectID) (
	img *Image, err error) {

	coll := db.Images()
	img = &Image{}

	err = coll.FindOne(db, &bson.M{
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

	err = coll.FindOne(db, &bson.M{
		"_id":          imgId,
		"organization": Global,
	}).Decode(img)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetOne(db *database.Database, query *bson.M) (img *Image, err error) {
	coll := db.Images()
	img = &Image{}

	err = coll.FindOne(db, query).Decode(img)
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

	keysInf, err := coll.Distinct(db, "key", &bson.M{
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

	n, err := coll.CountDocuments(db, &bson.M{
		"_id":          imgId,
		"organization": Global,
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

func GetAll(db *database.Database, query *bson.M) (
	imgs []*Image, err error) {

	coll := db.Images()
	imgs = []*Image{}

	cursor, err := coll.Find(db, query)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		img := &Image{}
		err = cursor.Decode(img)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		imgs = append(imgs, img)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllPaged(db *database.Database, query *bson.M, page, pageCount int64) (
	imgs []*Image, count int64, err error) {

	coll := db.Images()
	imgs = []*Image{}

	if len(*query) == 0 {
		count, err = coll.EstimatedDocumentCount(db)
		if err != nil {
			err = database.ParseError(err)
			return
		}
	} else {
		count, err = coll.CountDocuments(db, query)
		if err != nil {
			err = database.ParseError(err)
			return
		}
	}

	maxPage := count / pageCount
	if count == pageCount {
		maxPage = 0
	}
	page = utils.Min64(page, maxPage)
	skip := utils.Min64(page*pageCount, count)

	cursor, err := coll.Find(
		db,
		query,
		&options.FindOptions{
			Sort: &bson.D{
				{"key", 1},
			},
			Skip:  &skip,
			Limit: &pageCount,
		},
	)
	defer cursor.Close(db)

	for cursor.Next(db) {
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
		db,
		query,
		&options.FindOptions{
			Sort: &bson.D{
				{"key", 1},
			},
			Projection: &bson.D{
				{"name", 1},
				{"key", 1},
				{"signed", 1},
				{"firmware", 1},
			},
		},
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		img := &Image{}
		err = cursor.Decode(img)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		img.Json()

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
		db,
		&bson.M{},
		&options.FindOptions{
			Sort: &bson.D{
				{"key", 1},
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
	defer cursor.Close(db)

	for cursor.Next(db) {
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

	_, err = coll.DeleteOne(db, &bson.M{
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

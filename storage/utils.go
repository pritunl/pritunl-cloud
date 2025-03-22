package storage

import (
	"strings"

	minio "github.com/minio/minio-go/v7"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/mongo-go-driver/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/utils"
)

func Get(db *database.Database, storeId primitive.ObjectID) (
	store *Storage, err error) {

	coll := db.Storages()
	store = &Storage{}

	err = coll.FindOneId(storeId, store)
	if err != nil {
		return
	}

	return
}

func GetAll(db *database.Database) (stores []*Storage, err error) {
	coll := db.Storages()
	stores = []*Storage{}

	cursor, err := coll.Find(
		db,
		&bson.M{},
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		store := &Storage{}
		err = cursor.Decode(store)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		stores = append(stores, store)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllPaged(db *database.Database, query *bson.M,
	page, pageCount int64) (stores []*Storage, count int64, err error) {

	coll := db.Storages()
	stores = []*Storage{}

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
				{"name", 1},
			},
			Skip:  &skip,
			Limit: &pageCount,
		},
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		store := &Storage{}
		err = cursor.Decode(store)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		stores = append(stores, store)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func Remove(db *database.Database, storeId primitive.ObjectID) (err error) {
	coll := db.Images()

	_, err = coll.DeleteMany(db, &bson.M{
		"storage": storeId,
	})
	if err != nil {
		return
	}

	coll = db.Storages()

	_, err = coll.DeleteOne(db, &bson.M{
		"_id": storeId,
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

func RemoveMulti(db *database.Database, storeIds []primitive.ObjectID) (
	err error) {

	coll := db.Images()

	for _, storeId := range storeIds {
		_, err = coll.DeleteMany(db, &bson.M{
			"storage": storeId,
		})
		if err != nil {
			return
		}
	}

	coll = db.Storages()

	_, err = coll.DeleteMany(db, &bson.M{
		"_id": &bson.M{
			"$in": storeIds,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func FormatStorageClass(class string) string {
	switch class {
	case AwsStandard:
		return "STANDARD"
	case AwsInfrequentAccess:
		return "STANDARD_IA"
	case AwsGlacier:
		return "GLACIER"
	}

	return ""
}

func ParseStorageClass(obj minio.ObjectInfo) string {
	opcRequestId := obj.Metadata.Get("Opc-Request-Id")
	archivalState := strings.ToLower(obj.Metadata.Get("Archival-State"))
	if archivalState != "" {
		return OracleArchive
	} else if opcRequestId != "" {
		return OracleStandard
	}

	switch obj.StorageClass {
	case "STANDARD":
		return AwsStandard
	case "STANDARD_IA":
		return AwsInfrequentAccess
	case "GLACIER":
		return AwsGlacier
	}

	return ""
}

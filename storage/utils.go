package storage

import (
	"strings"

	minio "github.com/minio/minio-go"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
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

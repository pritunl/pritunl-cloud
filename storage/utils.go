package storage

import (
	"github.com/minio/minio-go"
	"github.com/pritunl/pritunl-cloud/database"
	"gopkg.in/mgo.v2/bson"
	"strings"
)

func Get(db *database.Database, storeId bson.ObjectId) (
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

	cursor := coll.Find(bson.M{}).Iter()

	nde := &Storage{}
	for cursor.Next(nde) {
		stores = append(stores, nde)
		nde = &Storage{}
	}

	err = cursor.Close()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func Remove(db *database.Database, storeId bson.ObjectId) (err error) {
	coll := db.Images()

	_, err = coll.RemoveAll(&bson.M{
		"storage": storeId,
	})
	if err != nil {
		return
	}

	coll = db.Storages()

	err = coll.Remove(&bson.M{
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

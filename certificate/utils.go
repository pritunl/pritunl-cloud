package certificate

import (
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/mongo-go-driver/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/utils"
)

func Get(db *database.Database, certId primitive.ObjectID) (
	cert *Certificate, err error) {

	coll := db.Certificates()
	cert = &Certificate{}

	err = coll.FindOneId(certId, cert)
	if err != nil {
		return
	}

	return
}

func GetOrg(db *database.Database, orgId, certId primitive.ObjectID) (
	cert *Certificate, err error) {

	coll := db.Certificates()
	cert = &Certificate{}

	err = coll.FindOne(db, &bson.M{
		"_id":          certId,
		"organization": orgId,
	}).Decode(cert)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetOne(db *database.Database, query *bson.M) (cert *Certificate, err error) {
	coll := db.Certificates()
	cert = &Certificate{}

	err = coll.FindOne(db, query).Decode(cert)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAll(db *database.Database,
	query *bson.M) (certs []*Certificate, err error) {

	coll := db.Certificates()
	certs = []*Certificate{}

	cursor, err := coll.Find(db, query)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		cert := &Certificate{}
		err = cursor.Decode(cert)
		if err != nil {
			return
		}

		certs = append(certs, cert)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllOrg(db *database.Database, orgId primitive.ObjectID) (
	certs []*Certificate, err error) {

	coll := db.Certificates()
	certs = []*Certificate{}

	cursor, err := coll.Find(db, &bson.M{
		"organization": orgId,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		cert := &Certificate{}
		err = cursor.Decode(cert)
		if err != nil {
			return
		}

		certs = append(certs, cert)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllPaged(db *database.Database, query *bson.M,
	page, pageCount int64) (certs []*Certificate, count int64, err error) {

	coll := db.Certificates()
	certs = []*Certificate{}

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
		cert := &Certificate{}
		err = cursor.Decode(cert)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		certs = append(certs, cert)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func Remove(db *database.Database, certId primitive.ObjectID) (err error) {
	coll := db.Certificates()

	_, err = coll.DeleteMany(db, &bson.M{
		"_id": certId,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func RemoveOrg(db *database.Database, orgId, certId primitive.ObjectID) (
	err error) {

	coll := db.Certificates()

	_, err = coll.DeleteOne(db, &bson.M{
		"_id":          certId,
		"organization": orgId,
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

func RemoveMulti(db *database.Database, certIds []primitive.ObjectID) (
	err error) {
	coll := db.Certificates()

	_, err = coll.DeleteMany(db, &bson.M{
		"_id": &bson.M{
			"$in": certIds,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func RemoveMultiOrg(db *database.Database, orgId primitive.ObjectID,
	certIds []primitive.ObjectID) (err error) {

	coll := db.Certificates()

	_, err = coll.DeleteMany(db, &bson.M{
		"_id": &bson.M{
			"$in": certIds,
		},
		"organization": orgId,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

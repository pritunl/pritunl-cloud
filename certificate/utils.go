package certificate

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/mongo-go-driver/v2/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/utils"
)

func Get(db *database.Database, certId bson.ObjectID) (
	cert *Certificate, err error) {

	coll := db.Certificates()
	cert = &Certificate{}

	err = coll.FindOneId(certId, cert)
	if err != nil {
		return
	}

	return
}

func GetOrg(db *database.Database, orgId, certId bson.ObjectID) (
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

func GetAllOrg(db *database.Database, orgId bson.ObjectID) (
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

func GetAllNames(db *database.Database, query *bson.M) (
	certs []*database.Named, err error) {

	coll := db.Certificates()
	certs = []*database.Named{}

	cursor, err := coll.Find(
		db,
		query,
		options.Find().
			SetSort(bson.D{{"name", 1}}).
			SetProjection(bson.D{{"name", 1}}),
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		crt := &database.Named{}
		err = cursor.Decode(crt)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		certs = append(certs, crt)
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

	if pageCount == 0 {
		pageCount = 20
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
		options.Find().
			SetSort(bson.D{{"name", 1}}).
			SetSkip(skip).
			SetLimit(pageCount),
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

func Remove(db *database.Database, certId bson.ObjectID) (err error) {
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

func RemoveOrg(db *database.Database, orgId, certId bson.ObjectID) (
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

func RemoveMulti(db *database.Database, certIds []bson.ObjectID) (
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

func RemoveMultiOrg(db *database.Database, orgId bson.ObjectID,
	certIds []bson.ObjectID) (err error) {

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
